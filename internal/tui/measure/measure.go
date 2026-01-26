package measure

type Window struct {
	Width  int
	Height int
}

func (w Window) ModalWidth() int {
	if w.Width >= 120 {
		return 80
	}
	if w.Width >= 80 {
		return w.Width - 20
	}
	return w.Width - 4
}

func (w Window) ModalHeight() int {
	return w.Height - 4
}

func (w Window) ListWidth() int {
	return w.ModalWidth() - 2
}

func (w Window) ListHeight() int {
	height := w.ModalHeight() - 6
	if height < 4 {
		height = 4
	}
	return height
}

func (w Window) ModalMargin() int {
	margin := (w.Width - w.ModalWidth()) / 2
	if margin < 1 {
		margin = 1
	}
	return margin
}

func (w Window) ContentWidth() int {
	return w.ModalWidth() - 4
}

func (w Window) IsSmall() bool {
	return w.Width < 60 || w.Height < 20
}
