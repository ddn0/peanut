package git

type Merge struct {
	Base    *Commit // Base commit for merge
	Topic   *Commit // Other commit to merge
	Current *Commit // Current commit
}

// Can Current merge Topic? i.e., can we git merge --ff-only Topic
func (a *Merge) CanFFMerge() bool {
	return a.Topic.Sha != a.Current.Sha && a.Base.Sha == a.Current.Sha
}
