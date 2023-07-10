
package set

import (
	"sync"
	"errors"
	"strings"
	"fmt"
	"strconv"
)

type uint64set struct {
	chunks []uint64
}