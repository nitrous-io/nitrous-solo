package data

import "golang.org/x/crypto/bcrypt"

type tab struct {
	Path           string `json:"path"`
	ScrollPosition struct {
		Left int64 `json:"left"`
		Top  int64 `json:"top"`
	} `json:"scroll_postion"`
	Selections []struct {
		Anchor struct {
			Line    int64 `json:"line"`
			Ch      int64 `json:"ch"`
			XRel    int64 `json:"xrel"`
			Outside bool  `json:"outside"`
		} `json:"anchor"`
		Head struct {
			Line    int64 `json:"line"`
			Ch      int64 `json:"ch"`
			XRel    int64 `json:"xrel"`
			Outside bool  `json:"outside"`
		} `json:"head"`
	} `json:"selections"`
	Active bool `json:"active"`
}

type UserData struct {
	Login                          string `json:"login"` // ide login username
	EncryptedPassword              []byte // ide login password
	User                           string `json:"user"` // username on system
	AutoCloseBracket               bool   `json:"auto_close_brackets"`
	FontSize                       int64  `json:"font_size"`
	HighlightLine                  bool   `json:"highlight_line"`
	Keybindings                    string `json:"keybindings"`
	MatchBrackets                  bool   `json:"match_brackets"`
	RemoveTrailingWhitespaceOnSave bool   `json:"remove_trailing_whitespace_on_save"`
	RulerColumn                    int64  `json:"ruler_column"`
	ShowLineNumbers                bool   `json:"show_line_numbers"`
	ShowTrailingSpace              bool   `json:"show_trailing_space"`
	SoftTabs                       bool   `json:"soft_tabs"`
	TabSize                        int64  `json:"tab_size"`
	Theme                          string `json:"theme"`
	WrapLines                      bool   `json:"wrap_lines"`
	RootPath                       string `json:"root_path"`
	HomePath                       string `json:"home_path"`
	ReadmeURL                      string `json:"readme_url"`
	ReadmeOpened                   bool   `json:"readme_opened"`
	OpenTabs                       []tab  `json:"open_tabs"`
}

func (d *UserData) SetPassword(pw string) error {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), 0)
	if err != nil {
		return err
	}
	d.EncryptedPassword = b
	return nil
}

func (d *UserData) VerifyPassword(pw string) bool {
	err := bcrypt.CompareHashAndPassword(d.EncryptedPassword, []byte(pw))
	return err == nil
}

func (d *UserData) SetDefaultSettings() {
	d.AutoCloseBracket = true
	d.FontSize = int64(14)
	d.HighlightLine = true
	d.Keybindings = "default"
	d.MatchBrackets = true
	d.RemoveTrailingWhitespaceOnSave = true
	d.RulerColumn = int64(0)
	d.ShowLineNumbers = true
	d.ShowTrailingSpace = false
	d.SoftTabs = true
	d.TabSize = int64(2)
	d.Theme = "solarized dark"
	d.WrapLines = true
	d.RootPath = "/home/nitrous"
	d.HomePath = "/home/nitrous"
	// FIXME: use nitrous readme
	d.ReadmeURL = "https://nitrous-readmes.s3.amazonaws.com/ruby.2015-06-29.html"
	d.ReadmeOpened = false
	d.OpenTabs = []tab{}
}

func (d *UserData) AsJSON() interface{} {
	return struct {
		Login                          string `json:"login"`
		User                           string `json:"user"`
		AutoCloseBracket               bool   `json:"auto_close_brackets"`
		FontSize                       int64  `json:"font_size"`
		HighlightLine                  bool   `json:"highlight_line"`
		Keybindings                    string `json:"keybindings"`
		MatchBrackets                  bool   `json:"match_brackets"`
		RemoveTrailingWhitespaceOnSave bool   `json:"remove_trailing_whitespace_on_save"`
		RulerColumn                    int64  `json:"ruler_column"`
		ShowLineNumbers                bool   `json:"show_line_numbers"`
		ShowTrailingSpace              bool   `json:"show_trailing_space"`
		SoftTabs                       bool   `json:"soft_tabs"`
		TabSize                        int64  `json:"tab_size"`
		Theme                          string `json:"theme"`
		WrapLines                      bool   `json:"wrap_lines"`
		RootPath                       string `json:"root_path"`
		HomePath                       string `json:"home_path"`
		ReadmeURL                      string `json:"readme_url"`
		ReadmeOpened                   bool   `json:"readme_opened"`
		OpenTabs                       []tab  `json:"open_tabs"`
	}{
		d.Login,
		d.User,
		d.AutoCloseBracket,
		d.FontSize,
		d.HighlightLine,
		d.Keybindings,
		d.MatchBrackets,
		d.RemoveTrailingWhitespaceOnSave,
		d.RulerColumn,
		d.ShowLineNumbers,
		d.ShowTrailingSpace,
		d.SoftTabs,
		d.TabSize,
		d.Theme,
		d.WrapLines,
		d.RootPath,
		d.HomePath,
		d.ReadmeURL,
		d.ReadmeOpened,
		d.OpenTabs,
	}
}
