package model

import (
	"fmt"
)

type Zone struct {
	A    *Record
	AAAA *Record
}

func (z Zone) String() string {
	ret := ""

	if z.A != nil {
		ret = ret + fmt.Sprintf("{A: %s, ", z.A.Target)
	} else {
		ret = ret + "{A: nil, "
	}

	if z.AAAA != nil {
		ret = ret + fmt.Sprintf("AAAA: %s}", z.AAAA.Target)
	} else {
		ret = ret + "AAAA: nil}"
	}

	return ret
}
