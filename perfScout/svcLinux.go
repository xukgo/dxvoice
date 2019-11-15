// +build linux

package perfScout

import (
	"github.com/shirou/gopsutil/net"
	"os"
	"runtime"
	"time"
)

var singleton *perfInfoMan

func Start() {
	singleton = new(perfInfoMan)
	go loopUpdateMonitorInfo()
}

func loopUpdateMonitorInfo() {
	firstRun := true
	hostName, _ := os.Hostname()
	singleton.setHostName(hostName)
	cpuCount := runtime.NumCPU()
	pid := os.Getpid()

	var lastTotalCpuTime int64
	var lastProCpuTime int64
	var lastIoInfo iorwInfo

	for {
		totalCpuTime, readTotalCpuErr := getTotalCpuTime()
		procCpuTime, readCpuErr := getProcCpuTime(pid)

		mem, readMemErr := getProcessMem(pid)
		fdSum, readFdErr := getProcessFdCount(pid)
		ioInfo, readIoErr := getIOBytes(pid)
		ninfo, readNetErr := net.ConnectionsPid("all", (int32(pid)))

		if !firstRun {
			singleton.locker.Lock()

			if readCpuErr == nil {
				loadrate := float64(cpuCount*100) * float64(procCpuTime-lastProCpuTime) / float64(totalCpuTime-lastTotalCpuTime)
				singleton.CpuLoadPercent = loadrate
			}
			if readMemErr == nil {
				singleton.MemUse = mem
			}
			if readFdErr == nil {
				singleton.FdCount = fdSum
			}
			if readIoErr == nil {
				singleton.WriteIOSpeed = float64(ioInfo.writeBytes-lastIoInfo.writeBytes) / float64(ioInfo.nanoStamp-lastIoInfo.nanoStamp)
			}

			if readNetErr == nil {
				singleton.SocketCount = len(ninfo)
			}

			singleton.locker.Unlock()
			//fmt.Printf("systemInfo:%v\r\n",singleton.PerfInfo)
		}
		firstRun = false

		if readCpuErr == nil {
			lastProCpuTime = procCpuTime
			lastTotalCpuTime = totalCpuTime
		}

		time.Sleep(time.Second * 2)

		//procInfo,_ := process.
		//ninfo,_ := net.ConnectionsPid("all", (int32(pid)))
		//fmt.Print(len(ninfo))

		//获取的是系统的总信息
		//cpuLoadPercents,err := cpu.Percent(time.Second*2,true)
		//if err == nil {
		//	singleton.SetCpuLoadPercent(cpuLoadPercents)
		//}
		//v, _ := mem.VirtualMemory()
		//d, _ := disk.Usage("/")
		//singleton.SetHostName(h.Hostname)
		//singleton.SetMemInfo((float64(v.Total))/1024/1024,(float64(v.Available))/1024/1024)
		//singleton.SetDiskInfo((float64(d.Total))/1024/1024,(float64(d.Free))/1024/1024)
	}
}