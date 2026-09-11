package cleanup

type Scheduler interface{
	start()
	stop()
	getStatus()
	scheduleNext()
	tick()
}