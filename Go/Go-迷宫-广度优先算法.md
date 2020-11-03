
由于要开一个爬虫项目，使用go语言实现，所以要利用广度优先算法，就打个基础，利用go语言实现这个算法

下面是代码实现：

<!--more-->

	package main
    
    import (
        "fmt"
        "os"
    )

    // 定义坐标数据结构
    type point struct {
        i, j int
    }

    // 定义上 左 下 右 是个方向需要相加的坐标
    var dirs = [4]point{
        {-1, 0},
        {0, -1},
        {1, 0},
        {0, 1},
    }
    // 定义坐标移动算法
    func (p point) add (r point) point{
        return point{p.i + r.i,p.j + r.j}
    }
    /**
        计算当前点点坐标值，还有是否越界
     */
    func (p point) at (grid [][]int)(int,bool){
        if p.i < 0 || p.i >= len(grid){
            return 0,false
        }

        if p.j < 0 || p.j >= len(grid[0]){
            return 0, false
        }

        return grid[p.i][p.j],true
    }
    // 读取迷宫文件
    // 此处默认文件一定存在，并且文件格式正确 例如下图
    // 第一行是迷宫行数，和列数
    
![upload successful](http://blogs.xinghe.host/images/pasted-46.png)
    
    func readMaze(filename string)[][]int{
        file, err := os.Open(filename)
        if err != nil {
            panic(err)
        }
        var col, row int
        // 读取第一行
        fmt.Fscanf(file, "%d %d", &row, &col)

        maze := make([][]int, row)
        for i := range maze {
            maze[i] = make([]int, col)
            for j := range maze[i] {
                fmt.Fscanf(file, "%d", &maze[i][j])
            }
        }

        return maze
    }
    func main (){
    	// 读取文件
        maze := readMaze("maze.txt")
        steps := walk(maze,point{0,0},point{len(maze)-1,len(maze[0])-1})

        for _,row := range steps{
            for _,v := range row{
                fmt.Printf("%3d",v)
            }
            fmt.Println()
        }
    }
    func walk(maze [][]int,start,end point)[][]int{
        // 新建答案数组
        steps := make([][]int, len(maze))
        for i, _ := range steps {
            steps[i] = make([]int, len(maze[i]))
        }
        // 新建队列，用来处理需要探索的点,首先添加起点
        queue := []point{start}
        // 当队列为空当时候 退出循环
        for len(queue) > 0 {
            // 当前需要探索的点
            cur := queue[0]
            // 把当前点取出队列
            queue = queue[1:]
            // 当走到终点当时候，退出
            if cur == end {
                break
            }
            // 循环探索当前点点四个方向是否符合要求，符合要求添加到steps中，并记录当前走到步数是第几步
            for _,dir := range dirs{
                next := cur.add(dir)
                // 查看当前探索点的下个方向是否越界或者是否是墙（值为1）
                val,ok := next.at(maze)
                if !ok || val == 1{
                    // 不满足 探索下个方向
                    continue
                }
                // 查看当前点属否在已走过点数组，已走过点的值（也就是第几步）肯定是大于0的
                val,ok = next.at(steps)
                if !ok || val != 0{
                    continue
                }
                // 回到起点也不行
                if next == start {
                    continue
                }
                // 满足条件，记录当前步数
                curStep,_ := cur.at(steps)
                // 记录到答案数组里面，其实是记录第几步了
                steps[next.i][next.j] = curStep + 1
                // 添加下个点到探索队列里面，用于探索下个点的四个方向
                queue = append(queue,next)
            }
        }

        return steps
    }

    