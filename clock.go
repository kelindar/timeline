package timeline

/*


type clock struct {
	interval  time.Duration
	scheduler *Scheduler
	ticker    *time.Ticker
}

func (c *clock) Start(ctx context.Context) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	// Calculate the time until the next 10ms boundary
	now := time.Now()
	nextTick := now.Truncate(c.interval).Add(c.interval)
	initialDelay := nextTick.Sub(now)

	// Wait until the next resolution boundary
	time.Sleep(initialDelay)

	// Start the ticker
	c.ticker = time.NewTicker(c.interval)
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.scheduler.Tick()
			case <-ctx.Done():
				c.ticker.Stop()
				return
			}
		}
	}()

	return cancel
}



func main() {
	cq := New()

	// Schedule some events
	now := time.Now()
	event1 := Event{Data: "Immediate Event 1"}
	event2 := Event{Data: "Immediate Event 2"}
	event3 := Event{Data: "Future Event 1"}
	event4 := Event{Data: "Future Event 2"}

	cq.Schedule(event1, now)
	cq.Schedule(event2, now.Add(5*time.Millisecond))
	cq.Schedule(event3, now.Add(495*time.Millisecond))
	cq.Schedule(event4, now.Add(1600*time.Millisecond))

	// Cancel event2
	cq.Cancel(&event2)

	fmt.Println("Processing events...")

	// Process events every 10ms (simulating the tick)
	ticker := time.NewTicker(initialWidth)
	count := 0
	for range ticker.C {
		fmt.Println("Tick", count)
		cq.ProcessCurrentBucket()

		// For demonstration purposes, stop after 2 seconds
		count++
		if count == 20 {
			ticker.Stop()
			break
		}
	}

	fmt.Println("Done")
}
*/
