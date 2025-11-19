package metrics

// IncrementProjectCreated increments project creation counter
func (m *Metrics) IncrementProjectCreated() {
	m.ProjectCreatedTotal.Inc()
}

// IncrementBoardCreated increments board creation counter
func (m *Metrics) IncrementBoardCreated() {
	m.BoardCreatedTotal.Inc()
}

// SetProjectsTotal sets total projects gauge
func (m *Metrics) SetProjectsTotal(count int64) {
	m.ProjectsTotal.Set(float64(count))
}

// SetBoardsTotal sets total boards gauge
func (m *Metrics) SetBoardsTotal(count int64) {
	m.BoardsTotal.Set(float64(count))
}
