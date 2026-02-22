package module

const (
	StatusPending    = "pending"
	StatusInstalling = "installing"
	StatusCompleted  = "completed"
	StatusError      = "error"
)

type Order struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Order   []string `json:"order"`
}

type Step struct {
	Type     string `json:"type"`
	File     string `json:"file,omitempty"`
	Args     string `json:"args,omitempty"`
	Command  string `json:"command,omitempty"`
	Variable string `json:"variable,omitempty"`
	Value    string `json:"value,omitempty"`
	Action   string `json:"action,omitempty"`
	Target   string `json:"target,omitempty"`
	Content  string `json:"content,omitempty"`
	Key      string `json:"key,omitempty"`
	Dest     string `json:"dest,omitempty"`
}

type Command struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
	Steps       []Step `json:"steps"`
}

type Module struct {
	FolderName string
	Command    Command
	Status     string
	Error      string
}

type ModuleStatus struct {
	FolderName  string `json:"folderName"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Weight      int    `json:"weight"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

func (m *Module) ToStatus() ModuleStatus {
	return ModuleStatus{
		FolderName:  m.FolderName,
		Name:        m.Command.Name,
		Description: m.Command.Description,
		Weight:      m.Command.Weight,
		Status:      m.Status,
		Error:       m.Error,
	}
}
