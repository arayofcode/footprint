package domain

type AssetKey string

func UserAvatarKey(username string) AssetKey {
	return AssetKey("user:" + username)
}

func RepoAvatarKey(repo string) AssetKey {
	return AssetKey("repo:" + repo)
}
