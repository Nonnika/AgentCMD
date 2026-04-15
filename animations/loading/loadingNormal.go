// LoadingNormal 正常加载动画
package loading

import (
	"fmt"
	"io"
	"time"
)

// 不需要 defer close(done) 因为我们使用的是通道的关闭信号来判断是否需要停止加载动画
func LoadingNormal(done chan struct{}, headString string, doneString string, output io.Writer) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for {
		for _, frame := range frames {
			select {
			case <-done:
				fmt.Fprintf(output, "\033[2K\r\033[34m%s\033[0m", doneString)
				return
			default:
				fmt.Fprintf(output, "\r\033[34m%s %s\033[0m", headString, frame)
				time.Sleep(100 * time.Millisecond)
			}
		}

	}
}
