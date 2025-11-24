package utils

import (
	"crypto/rand"
	"fmt"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const digits = "0123456789"

func randomCharFromSet(set string) byte {
	n := make([]byte, 1)
	_, err := rand.Read(n)
	if err != nil {
		panic(err)
	}
	return set[int(n[0])%len(set)]
}

// TODO: this does not guarantee unique ids every time..
func GenerateBucketKey() (string, error) {
	part1 := make([]byte, 3)
	part2 := make([]byte, 3)
	part3 := make([]byte, 2)

	for i := range 3 {
		part1[i] = randomCharFromSet(letters)
		part2[i] = randomCharFromSet(letters)
	}
	for i := range 2 {
		part3[i] = randomCharFromSet(digits)
	}

	return fmt.Sprintf("%s-%s-%s", part1, part2, part3), nil
}
