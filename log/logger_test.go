package log

import (
	"log"
	"testing"

	"emperror.dev/errors"
)

func TestName(t *testing.T) {
	t.Log(IsBackgroundMode())
}

func lowLevel() error {
	return errors.New("root cause at low level")
}

func middle() error {
	if err := lowLevel(); err != nil {
		return errors.WithStackIf(err)
	}
	return nil
}

func TestName2(t *testing.T) {
	if err := middle(); err != nil {
		//fmt.Printf("Detailed Error: %+v\n", err)
		log.Printf("%+v\n", err)
	}
}
