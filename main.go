package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ANSI 颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func header(msg string) {
	fmt.Println(ColorCyan + msg + ColorReset)
}

func info(format string, a ...interface{}) {
	fmt.Printf(ColorWhite+format+ColorReset, a...)
}

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
	hostname, _ := os.Hostname()

	// 读取内核版本
	kernelVer := "unknown"
	if b, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		kernelVer = strings.TrimSpace(string(b))
	}

	// 读取启动时间 /proc/uptime
	var uptimeSeconds float64
	if b, err := os.ReadFile("/proc/uptime"); err == nil {
		parts := strings.Fields(string(b))
		if len(parts) > 0 {
			uptimeSeconds, _ = strconv.ParseFloat(parts[0], 64)
		}
	}
	bootTime := time.Now().Add(-time.Duration(uptimeSeconds) * time.Second)

	info("  主机名:   %s\n", hostname)
	info("  OS:       %s %s\n", runtime.GOOS, runtime.GOARCH)
	info("  内核:     %s\n", kernelVer)
	info("  架构:     %s\n", runtime.GOARCH)
	info("  启动时间: %s (已运行 %d 小时)\n", bootTime.Format("2006-01-02 15:04"), int(uptimeSeconds)/3600)
}

// 2. CPU 信息
func printCPUInfo() {
	header("\n[2] CPU 状态")

	// 获取 CPU 型号
	modelName := "Unknown"
	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "model name") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					modelName = strings.TrimSpace(parts[1])
					break
				}
			}
		}
	}

	info("  型号:     %s\n", modelName)
	info("  核心数:   %d Cores\n", runtime.NumCPU())

	// 计算 CPU 使用率 (采样 /proc/stat)
	idle1, total1 := getCPUSample()
	time.Sleep(200 * time.Millisecond) // 采样间隔
	idle2, total2 := getCPUSample()

	idleTicks := float64(idle2 - idle1)
	totalTicks := float64(total2 - total1)
	cpuUsage := 0.0
	if totalTicks > 0 {
		cpuUsage = 100 * (totalTicks - idleTicks) / totalTicks
	}

	printBar("CPU 使用率", cpuUsage)
}

func getCPUSample() (idle, total uint64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) > 4 && fields[0] == "cpu" {
			// user, nice, system, idle, iowait, irq, softirq, steal
			for i, v := range fields[1:] {
				val, _ := strconv.ParseUint(v, 10, 64)
				total += val
				if i == 3 { // idle is the 4th field (index 3)
					idle = val
				}
			}
		}
	}
	return
}

// 3. 内存信息
func printMemInfo() {
	header("\n[3] 内存状态")

	var memTotal, memAvailable uint64
	if f, err := os.Open("/proc/meminfo"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			key := parts[0]
			val, _ := strconv.ParseUint(parts[1], 10, 64) // kB

			if strings.HasPrefix(key, "MemTotal") {
				memTotal = val * 1024
			} else if strings.HasPrefix(key, "MemAvailable") {
				memAvailable = val * 1024
			}
		}
	}

	totalGB := float64(memTotal) / 1024 / 1024 / 1024
	usedPercent := 0.0
	if memTotal > 0 {
		usedPercent = float64(memTotal-memAvailable) / float64(memTotal) * 100
	}

	info("  总内存:   %.2f GB\n", totalGB)
	printBar("内存使用率", usedPercent)
}

// 4. 磁盘信息 (只看根目录 / )
func printDiskInfo() {
	header("\n[4] 磁盘空间 (根目录 /)")

	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		total := float64(stat.Blocks) * float64(stat.Bsize)
		free := float64(stat.Bavail) * float64(stat.Bsize)
		used := total - free

		totalGB := total / 1024 / 1024 / 1024
		freeGB := free / 1024 / 1024 / 1024
		usedPercent := 0.0
		if total > 0 {
			usedPercent = (used / total) * 100
		}

		info("  总量:     %.2f GB\n", totalGB)
		info("  剩余:     %.2f GB\n", freeGB)
		printBar("磁盘使用率", usedPercent)
	}
}

// 5. 网络端口
func printNetInfo() {
	header("\n[5] 关键端口监听")

	ports := make(map[int64]bool)

	// 读取 /proc/net/tcp 和 tcp6
	for _, file := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		defer f.Close() // 注意：在循环中使用 defer 会在函数结束时才关闭，这里文件少影响不大，严谨写法应封装函数

		scanner := bufio.NewScanner(f)
		scanner.Scan() // skip header
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 4 {
				continue
			}
			// state 0A is LISTEN
			if fields[3] == "0A" {
				// local_address is like 00000000:0016
				addrParts := strings.Split(fields[1], ":")
				if len(addrParts) == 2 {
					port, _ := strconv.ParseInt(addrParts[1], 16, 64)
					ports[port] = true
				}
			}
		}
	}

	info("  TCP 监听端口: ")
	for p := range ports {
		fmt.Printf(ColorGreen+"[%d] "+ColorReset, p)
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
	colorCode := ColorGreen
	if percent > 80 {
		colorCode = ColorRed
	}

	info("  %-10s: ", label)
	fmt.Printf("%s%s %.2f%%%s\n", colorCode, bar, percent, ColorReset)
}
