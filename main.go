package main

import (
	"encoding/json"
	"github.com/MarinX/keylogger"
	"github.com/jasonlvhit/gocron"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"openSelf/presenter"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)


func LogKeyboard(info *Info) {
	keyboard := keylogger.FindKeyboardDevice()

	// check if we found a path to keyboard
	if len(keyboard) <= 0 {
		logrus.Error("No keyboard found...you will need to provide manual input path")
		return
	}
	// init keylogger with keyboard
	k, err := keylogger.New(keyboard)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer k.Close()

	events := k.Read()


	for e := range events {
		switch e.Type {
		case keylogger.EvKey:

			// if the state of key is pressed
			if e.KeyPress() {
				info.ObserveKeyPressed()
			}
			break
		}
	}
}


func GetCurrentAppName() string {
	out, err := exec.Command("xdotool", "getwindowfocus", "getwindowpid").CombinedOutput()
	if err != nil {
		output := string(out[:])
		logrus.Error(output)
		panic(err)
	}
	output := string(out[:])
	content, err := ioutil.ReadFile("/proc/" + strings.TrimSpace(output) + "/comm")
	if err != nil {
		return "error"
	}
	text := string(content)
	return text
}


type AppUsage struct {
	Name string
	UsageInMin float64
}

type AppLog struct {
	Name string
	OpenedAt time.Time
	ClosedAt time.Time
}


type TypingInfo struct {
	TotalKeys int
	CreatedAt time.Time
	ClosedAt time.Time
}

type Info struct {
	CreatedAt time.Time
	AppsUsage []*AppUsage
	AppsLogs []*AppLog
	TypingInfo []*TypingInfo
	IsClosed bool
}

func (info *Info) ObserveKeyPressed() {
	if len(info.TypingInfo) == 0 {
		info.TypingInfo = append(info.TypingInfo, &TypingInfo{CreatedAt:time.Now()})
	}
	// Last element might be current typing info holder
	currentTypingInfo := info.TypingInfo[len(info.TypingInfo) - 1]
	if time.Now().Sub(currentTypingInfo.CreatedAt).Minutes() >= 1 {
		// Flush current typing info
		currentTypingInfo.ClosedAt = time.Now()
		currentTypingInfo = &TypingInfo{CreatedAt:time.Now()}
		info.TypingInfo = append(info.TypingInfo, currentTypingInfo)
	}
	currentTypingInfo.TotalKeys++
}


func GetAppUsage(info *Info, appName string) *AppUsage {
	for _, appUsage := range info.AppsUsage {
		if appUsage.Name == appName {
			return appUsage
		}
	}
	appUsage := &AppUsage{Name:appName}
	info.AppsUsage = append(info.AppsUsage, appUsage)
	return appUsage
}

func CalculateTimeUsage(info *Info) {
	for index := 0; index < len(info.AppsUsage); index++ {
		info.AppsUsage[index].UsageInMin = 0
	}
	for index := 0; index < len(info.AppsLogs); index++ {
		log := info.AppsLogs[index]
		if !log.ClosedAt.IsZero() {
			GetAppUsage(info, log.Name).UsageInMin += log.ClosedAt.Sub(log.OpenedAt).Minutes()
		}
	}
}

func CheckCurrentApp(info *Info)  {
	currentAppName := GetCurrentAppName()
	if len(info.AppsLogs) == 0 {
		info.AppsLogs = append(info.AppsLogs, &AppLog{Name:currentAppName, OpenedAt:time.Now()})
	}
	lastAppLog := info.AppsLogs[len(info.AppsLogs) - 1]
	if lastAppLog.Name != currentAppName {
		lastAppLog.ClosedAt = time.Now()
		info.AppsLogs = append(info.AppsLogs, &AppLog{Name:currentAppName, OpenedAt:time.Now()})
	}
	CalculateTimeUsage(info)
}

func PersistInfo(info *Info)  {
	fileName := GetInfoFile(info.CreatedAt)
	bytes, _ := json.Marshal(info)
	err := ioutil.WriteFile(fileName, bytes, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func CollectInfo(info *Info) {
	go LogKeyboard(info)
	scheduler := gocron.NewScheduler()
	CheckCurrentApp(info)
	scheduler.Every(10).Seconds().Do(CheckCurrentApp, info)
	scheduler.Every(1).Minutes().Do(PersistInfo, info)
	<- scheduler.Start()
}

func CloseInfo(info *Info)  {
	lastAppLog := info.AppsLogs[len(info.AppsLogs) - 1]
	lastAppLog.ClosedAt = time.Now()
	currentTypingInfo := info.TypingInfo[len(info.TypingInfo) - 1]
	currentTypingInfo.ClosedAt = time.Now()
	CalculateTimeUsage(info)
	PersistInfo(info)
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func GetNewInfo() *Info {
	return &Info{CreatedAt:time.Now()}
}

func GetInfoFile(t time.Time) string {
	os.MkdirAll(os.Getenv("OUTPUT_DIR"), os.ModePerm)
	return os.Getenv("OUTPUT_DIR") + t.Format("01-02-2006") + ".json"
}

func GetInfo() *Info {
	fileName := GetInfoFile(time.Now())
	if !Exists(fileName) {
		return GetNewInfo()
	}
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return GetNewInfo()
	}
	info := &Info{}
	err = json.Unmarshal(bytes, info)
	if err != nil {
		return GetNewInfo()
	}
	return info
}

func main() {
	presenter.StartServer()
	logrus.Info("OpenSelf service installed")
	info := GetInfo()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		CloseInfo(info)
		logrus.Info("OpenSelf service gracefully exited")
		os.Exit(1)
	}()
	CollectInfo(info)
}