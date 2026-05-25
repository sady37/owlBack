package measure

// MaxInscribedRect 在 polygon ∩ fov 的栅格里找最大轴对齐矩形（cm）。
// 算法：栅格化为 binary grid（cell 中心 in 两 polygon）→ histogram + monotonic stack。
func MaxInscribedRect(polygon Polygon, fov Polygon) (Rectangle, bool) {
	if len(polygon.Vertices) < 3 {
		return Rectangle{}, false
	}
	pb := polygonBBox(polygon)
	if len(fov.Vertices) >= 3 {
		fb := polygonBBox(fov)
		if fb.MinX > pb.MinX {
			pb.MinX = fb.MinX
		}
		if fb.MinY > pb.MinY {
			pb.MinY = fb.MinY
		}
		if fb.MaxX < pb.MaxX {
			pb.MaxX = fb.MaxX
		}
		if fb.MaxY < pb.MaxY {
			pb.MaxY = fb.MaxY
		}
	}
	if pb.MaxX <= pb.MinX || pb.MaxY <= pb.MinY {
		return Rectangle{}, false
	}
	cols := (pb.MaxX-pb.MinX)/CellSize + 1
	rows := (pb.MaxY-pb.MinY)/CellSize + 1
	grid := make([][]bool, rows)
	hasFOV := len(fov.Vertices) >= 3
	for r := 0; r < rows; r++ {
		grid[r] = make([]bool, cols)
		for c := 0; c < cols; c++ {
			p := Point{X: pb.MinX + c*CellSize + CellSize/2, Y: pb.MinY + r*CellSize + CellSize/2}
			if !pointInPolygon(p, polygon) {
				continue
			}
			if hasFOV && !pointInPolygon(p, fov) {
				continue
			}
			grid[r][c] = true
		}
	}
	col, row, w, h := maxRectInBinary(grid)
	if w == 0 || h == 0 {
		return Rectangle{}, false
	}
	return Rectangle{
		X:      pb.MinX + col*CellSize,
		Y:      pb.MinY + row*CellSize,
		Width:  w * CellSize,
		Height: h * CellSize,
	}, true
}

// maxRectInBinary 经典：rows × cols histogram + 单调栈 O(rows×cols)。返回 (col, row, w, h)（cell 单位）。
func maxRectInBinary(grid [][]bool) (int, int, int, int) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return 0, 0, 0, 0
	}
	rows := len(grid)
	cols := len(grid[0])
	heights := make([]int, cols+1)
	best := 0
	var bestCol, bestRow, bestW, bestH int
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if grid[r][c] {
				heights[c]++
			} else {
				heights[c] = 0
			}
		}
		stack := []int{}
		for c := 0; c <= cols; c++ {
			for len(stack) > 0 && heights[stack[len(stack)-1]] > heights[c] {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				left := -1
				if len(stack) > 0 {
					left = stack[len(stack)-1]
				}
				w := c - left - 1
				h := heights[top]
				area := w * h
				if area > best {
					best = area
					bestW = w
					bestH = h
					bestCol = left + 1
					bestRow = r - h + 1
				}
			}
			stack = append(stack, c)
		}
	}
	return bestCol, bestRow, bestW, bestH
}
