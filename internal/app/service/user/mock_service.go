package user

import "net/http"

type MockUserService struct {
	userID string
}

func NewMockUserService(userID string) *MockUserService {
	return &MockUserService{userID: userID}
}

func (m *MockUserService) GetUserIDFromCookie(r *http.Request) string {
	return m.userID
}

func (m *MockUserService) GenerateUserToken() (string, error) {
	return "mock-token", nil
}

func (m *MockUserService) VerifyUserToken(token string) (string, error) {
	return m.userID, nil
}
