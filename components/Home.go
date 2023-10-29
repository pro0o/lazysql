package components

import (
	"lazysql/app"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Home struct {
	*tview.Flex
	Tree           *Tree
	TabbedPane     *TabbedPane
	LeftWrapper    *tview.Flex
	RightWrapper   *tview.Flex
	FocusedWrapper string
}

func NewHomePage(name string) *Home {
	tree := NewTree()
	tabbedPane := NewTabbedPane()
	leftWrapper := tview.NewFlex()
	rightWrapper := tview.NewFlex()

	home := &Home{
		Flex:         tview.NewFlex(),
		Tree:         tree,
		TabbedPane:   tabbedPane,
		LeftWrapper:  leftWrapper,
		RightWrapper: rightWrapper,
	}

	go home.subscribeToTreeChanges()

	leftWrapper.SetBorderColor(app.InactiveTextColor)

	rightWrapper.SetBorderColor(app.InactiveTextColor)
	rightWrapper.SetBorder(true)
	rightWrapper.SetDirection(tview.FlexColumnCSS)

	rightWrapper.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var tab *Tab

		if event.Rune() == '[' {
			focusTab(tabbedPane.SwitchToPreviousTab())
			return nil
		} else if event.Rune() == ']' {
			focusTab(tabbedPane.SwitchToNextTab())
			return nil
		} else if event.Rune() == '{' {
			focusTab(tabbedPane.SwitchToFirstTab())
			return nil
		} else if event.Rune() == '}' {
			focusTab(tabbedPane.SwitchToLastTab())
			return nil
		} else if event.Rune() == 'X' {
			tab = tabbedPane.GetCurrentTab()

			if tab != nil {
				table := tab.Content

				if !table.GetIsFiltering() && !table.GetIsEditing() && !table.GetIsLoading() {
					tabbedPane.RemoveCurrentTab()

					if tabbedPane.GetLenght() == 0 {
						home.focusLeftWrapper()
						return nil
					}
				}
			}
		} else if event.Rune() == '<' {
			tab = tabbedPane.GetCurrentTab()

			if tab != nil {
				table := tab.Content

				if ((table.Menu != nil && table.Menu.GetSelectedOption() == 1) || table.Menu == nil) && !table.Pagination.GetIsFirstPage() && !table.GetIsLoading() {
					table.Pagination.SetOffset(table.Pagination.GetOffset() - table.Pagination.GetLimit())
					table.FetchRecords(table.GetDBReference())

				}

			}

		} else if event.Rune() == '>' {
			tab = tabbedPane.GetCurrentTab()

			if tab != nil {
				table := tab.Content

				if ((table.Menu != nil && table.Menu.GetSelectedOption() == 1) || table.Menu == nil) && !table.Pagination.GetIsLastPage() && !table.GetIsLoading() {
					table.Pagination.SetOffset(table.Pagination.GetOffset() + table.Pagination.GetLimit())
					table.FetchRecords(table.GetDBReference())
				}
			}
		}

		return event
	})
	leftWrapper.AddItem(tree, 0, 1, true)

	rightWrapper.AddItem(tabbedPane.HeaderContainer, 1, 0, false)
	rightWrapper.AddItem(tabbedPane.Pages, 0, 1, false)

	home.AddItem(leftWrapper, 30, 1, false)
	home.AddItem(rightWrapper, 0, 5, false)

	home.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		tab := tabbedPane.GetCurrentTab()

		var table *ResultsTable = nil

		if tab != nil {
			table = tab.Content
		}

		if event.Rune() == 'H' {
			if table != nil && !table.GetIsEditing() && !table.GetIsFiltering() && home.FocusedWrapper == "right" {
				home.focusLeftWrapper()
			}
		} else if event.Rune() == 'L' {
			if table != nil && !table.GetIsEditing() && !table.GetIsFiltering() && home.FocusedWrapper == "left" {
				home.focusRightWrapper()
			}
		} else if event.Rune() == 5 { // CTRL + e
			tab := tabbedPane.GetTabByName("Editor")

			if tab != nil {
				tabbedPane.SwitchToTabByName("Editor")
			} else {
				tableWithEditor := NewResultsTable().WithEditor()
				tabbedPane.AppendTab("Editor", tableWithEditor)
				tableWithEditor.SetIsFiltering(true)
			}
			home.focusRightWrapper()
			App.ForceDraw()
		} else if event.Rune() == 127 {
			if (table != nil && !table.GetIsEditing() && !table.GetIsFiltering() && !table.GetIsLoading()) || table == nil {
				MainPages.SwitchToPage("Connections")
			}
		} else if event.Rune() == 'q' {
			if tab != nil {
				table := tab.Content

				if !table.GetIsFiltering() && !table.GetIsEditing() {
					App.Stop()
				}
			} else {
				App.Stop()
			}
		}

		return event
	})

	home.SetFocusFunc(func() {
		if home.FocusedWrapper == "left" || home.FocusedWrapper == "" {
			home.focusLeftWrapper()
		} else {
			home.focusRightWrapper()
		}
	})

	MainPages.AddPage(name, home, true, false)
	return home
}

func (home *Home) subscribeToTreeChanges() {
	ch := home.Tree.Subscribe()

	for stateChange := range ch {
		switch stateChange.Key {
		case "SelectedTable":
			tableName := stateChange.Value.(string)

			tab := home.TabbedPane.GetTabByName(tableName)
			var table *ResultsTable = nil

			if tab != nil {
				table = tab.Content
				home.TabbedPane.SwitchToTabByName(tab.Name)
			} else {
				table = NewResultsTable().WithFilter()

				home.TabbedPane.AppendTab(tableName, table)
			}

			home.focusRightWrapper()

			table.FetchRecords(tableName)

			app.App.ForceDraw()
		}
	}
}

func (home *Home) focusRightWrapper() {
	home.Tree.RemoveHighlight()

	home.RightWrapper.SetBorderColor(app.FocusTextColor)
	home.LeftWrapper.SetBorderColor(app.InactiveTextColor)
	home.TabbedPane.Highlight()
	tab := home.TabbedPane.GetCurrentTab()

	if tab != nil {
		focusTab(tab)
	}

	home.FocusedWrapper = "right"
}

func focusTab(tab *Tab) {
	if tab != nil {
		table := tab.Content
		table.HighlightAll()

		if table.GetIsFiltering() {
			go func() {
				if table.Filter != nil {
					App.SetFocus(table.Filter.Input)
					table.Filter.HighlightLocal()
				} else if table.Editor != nil {
					App.SetFocus(table.Editor)
					table.Editor.Highlight()
				}

				table.RemoveHighlightTable()
				App.Draw()
			}()
		} else {
			App.SetFocus(table)
		}

	}
}

func (home *Home) focusLeftWrapper() {
	home.Tree.Highlight()

	home.RightWrapper.SetBorderColor(app.InactiveTextColor)
	home.LeftWrapper.SetBorderColor(app.FocusTextColor)

	tab := home.TabbedPane.GetCurrentTab()

	if tab != nil {
		table := tab.Content

		table.RemoveHighlightAll()

	}

	home.TabbedPane.SetBlur()

	App.SetFocus(home.Tree)

	home.FocusedWrapper = "left"
}