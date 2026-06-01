package fallback

const (
	ActionRetryNext = "retry_next"
	ActionFail      = "fail"
)

type Input struct {
	Attempt int
	Total   int
	Err     error
}

type Decision struct {
	Action      string
	Explanation string
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Decide(input Input) Decision {
	if input.Attempt+1 < input.Total {
		return Decision{
			Action:      ActionRetryNext,
			Explanation: "Primary route failed; trying next policy-allowed candidate",
		}
	}
	return Decision{
		Action:      ActionFail,
		Explanation: "No more policy-allowed fallback candidates",
	}
}
