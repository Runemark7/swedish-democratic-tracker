package postgres

// SetBallotsFirstMax lets tests force either ListByPolitician strategy.
func SetBallotsFirstMax(n int) (restore func()) {
	old := ballotsFirstMax
	ballotsFirstMax = n
	return func() { ballotsFirstMax = old }
}
