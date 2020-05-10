package release

import (
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/diff"
)

type fakeGit struct {
	input string
}

func (g fakeGit) exec(commands ...string) (string, error) {
	return g.input, nil
}

func Test_mergeLogForRepo(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name    string
		input   string
		repo    string
		from    string
		to      string
		want    []MergeCommit
		wantErr bool
	}{
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1eBug 1743564: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1eBug 1743564: test [trailing]",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test [trailing]",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e[release-4.1] Bug 1743564: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release-4.1] Bug 1743564: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release-4.1] Bug 1743564 : test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release-4.1] Bugs 1743564 : test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release-4.1] Bugs 1743564,: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{1743564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release-4.1] Bugs , 17 43,564,: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{17, 43, 564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release-4.1] bugs , 17 43,564,: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: []int{17, 43, 564}, Subject: "test",
				},
			},
		},
		{
			input: "abc\x1e1\x1eMerge pull request #145 from\x1e [release 4.1] bugs , 17 43,564,: test",
			want: []MergeCommit{
				{
					ParentCommits: []string{}, Commit: "abc", PullRequest: 145, CommitDate: time.Unix(1, 0).UTC(),
					Bugs: nil, Subject: "[release 4.1] bugs , 17 43,564,: test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := fakeGit{input: tt.input}
			got, err := mergeLogForRepo(g, tt.repo, "a", "b")
			if (err != nil) != tt.wantErr {
				t.Errorf("mergeLogForRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeLogForRepo(): %s", diff.ObjectReflectDiff(tt.want, got))
			}
		})
	}
}
