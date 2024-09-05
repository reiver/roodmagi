package blueskysrv

import (
	"fmt"
	"math/rand/v2"
	"time"
)

func chronorand() string {
	return fmt.Sprintf("id:?unix=%d&rand=0x%X", time.Now().Unix(), rand.Int64())
}
