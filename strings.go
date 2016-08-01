package main

import (
	"fmt"
)

type Strings []string

func (s *Strings) String() string {
	return fmt.Sprintf("%s", *s)
}

func (s *Strings) Set(value string) error {
	*s = append(*s, ip2resolver(value))
	return nil
}
