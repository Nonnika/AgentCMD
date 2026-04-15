// LoadingNormal 正常加载动画
package loading

import (
	"fmt"
	"time"
)

// 不需要 defer close(done) 因为我们使用的是通道的关闭信号来判断是否需要停止加载动画
func LoadingNormal(done chan struct{}, headString string, doneString string) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for {
		for _, frame := range frames {
			select {
			case <-done:
				fmt.Printf("\033[2K\r%s\n", doneString)
				return
			default:
				fmt.Printf("\r%s %s", headString, frame)
				time.Sleep(100 * time.Millisecond)
			}
		}

	}
}
