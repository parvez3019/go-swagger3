package handler

// live is the liveness handler.
// @Success 200 "live endpoint"
// @Router  /live [get]
func live() {}

// pushUpdate inserts a new update.
// @Accept  multipart/form-data
// @Success 201 {string} string
// @Router  /updates [post]
func update() {}
