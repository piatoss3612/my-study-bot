package admin

import (
	"context"
	"time"

	"github.com/piatoss3612/my-study-bot/internal/study"
)

// publish event to subscriber
func (ac *adminCommand) publishEvent(evt study.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cnt := 0

	for {
		select {
		case <-ctx.Done():
			ac.sugar.Errorw("failed to publish event", "error", ctx.Err().Error(), "topic", evt.Topic.String(), "description", evt.Description, "retry", cnt)
			return
		default:
			err := ac.pub.Publish(ctx, evt.Topic.String(), evt)
			if err != nil {
				ac.sugar.Errorw("failed to publish event", "error", err.Error(), "topic", evt.Topic.String(), "description", evt.Description, "retry", cnt)
				time.Sleep(500 * time.Millisecond)
				cnt++
				continue
			}
			ac.sugar.Infow("event published", "topic", evt.Topic.String(), "retry", cnt)
			return
		}
	}
}
