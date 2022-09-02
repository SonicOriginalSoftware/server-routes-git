//revive:disable:package-comments

package internal

const (
	// InfoPath is the info portion of an info/refs path
	InfoPath = "info"
	// RefsPath is the refs portion of an info/refs path
	RefsPath = "refs"

	// ReceivePackPath is the path for a receive pack request
	ReceivePackPath = "git-receive-pack"
	// UploadPackPath is the path for an upload pack request
	UploadPackPath = "git-upload-pack"
)
