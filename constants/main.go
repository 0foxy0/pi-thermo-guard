package constants

const TempPath = "/sys/devices/virtual/thermal/thermal_zone0/temp"

var ShutdownCmd = []string{"shutdown", "-h", "now"}

const EmailSubject = "❗️Raspberry Pi overheated"
const EmailFromName = "Raspberry Pi"
