package event

import (
	"EventTrigger/data"
	"log"
	"time"
)

func StartRetentionJob() {
	go func() {
		ticker := time.NewTicker(time.Hour * 1)
		for {
			select {
			case <-ticker.C:
				if err := data.DB.Where("timestamp < ?", time.Now().Add(-time.Hour*48)).Delete(&data.Event{}).Error; err != nil {
					log.Fatalf("Failed to delete logs before %v reson %v", time.Now().Add(-time.Hour*48), err)
				}
			}
		}
	}()
}
