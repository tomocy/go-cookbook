package usecase

type DoHaveAccessToken struct{}

func (u DoHaveAccessToken) Do(id string) (bool, error) {
	return id != "test_user_id", nil
}
