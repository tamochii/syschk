package main

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// 定义一些颜色打印函数
var (
	header = color.New(color.FgHiCyan, color.Bold).PrintlnFunc()
	info   = color.New(color.FgWhite).PrintfFunc()
	warn   = color.New(color.FgHiYellow).PrintfFunc()
	danger = color.New(color.FgHiRed).PrintfFunc()
)

func main() {

	printHostInfo()
	printCPUInfo()
	printMemInfo()
	printDiskInfo()
	printNetInfo()

}

// 1. 系统基础信息
func printHostInfo() {
	header("[1] 系统概况")
	h, _ := host.Info()
	info("  主机名:   %s\n", h.Hostname)
	info("  OS:       %s %s\n", h.Platform, h.PlatformVersion)
	info("  内核:     %s\n", h.KernelVersion)
	info("  架构:     %s\n", runtime.GOARCH)
	info("  启动时间: %s (已运行 %d 小时)\n", time.Unix(int64(h.BootTime), 0).Format("2006-01-02 15:04"), h.Uptime/3600)
}

// 2. CPU 信息
func printCPUInfo() {
	header("\n[2] CPU 状态")
	c, _ := cpu.Info()
	if len(c) > 0 {
		info("  型号:     %s\n", c[0].ModelName)
		info("  核心数:   %d Cores\n", len(c))
	}

	percent, _ := cpu.Percent(time.Second, false)
	if len(percent) > 0 {
		used := percent[0]
		printBar("CPU 使用率", used)
	}
}

// 3. 内存信息
func printMemInfo() {
	header("\n[3] 内存状态")
	v, _ := mem.VirtualMemory()

	totalGB := float64(v.Total) / 1024 / 1024 / 1024
	info("  总内存:   %.2f GB\n", totalGB)
	printBar("内存使用率", v.UsedPercent)
}

// 4. 磁盘信息 (只看根目录 / )
func printDiskInfo() {
	header("\n[4] 磁盘空间 (根目录 /)")
	d, err := disk.Usage("/")
	if err == nil {
		totalGB := float64(d.Total) / 1024 / 1024 / 1024
		freeGB := float64(d.Free) / 1024 / 1024 / 1024
		info("  总量:     %.2f GB\n", totalGB)
		info("  剩余:     %.2f GB\n", freeGB)
		printBar("磁盘使用率", d.UsedPercent)
	}
}

// 5. 网络端口
func printNetInfo() {
	header("\n[5] 关键端口监听")
	conns, _ := net.Connections("tcp")

	// 过滤去重
	ports := make(map[uint32]bool)
	for _, conn := range conns {
		if conn.Status == "LISTEN" {
			ports[conn.Laddr.Port] = true
		}
	}

	info("  TCP 监听端口: ")
	for p := range ports {
		color.New(color.FgHiGreen).Printf("[%d] ", p)
	}
	fmt.Println()
}

// 辅助函数：打印进度条
func printBar(label string, percent float64) {
	barLength := 20
	usedLength := int(math.Round(percent / 100 * float64(barLength)))

	bar := ""
	for i := 0; i < barLength; i++ {
		if i < usedLength {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	// 80% 以上标红
	c := color.New(color.FgGreen)
	if percent > 80 {
		c = color.New(color.FgRed, color.Bold)
	}

	info("  %-10s: ", label)
	c.Printf("%s %.2f%%\n", bar, percent)
}
