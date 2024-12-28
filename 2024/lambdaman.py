#!/usr/bin/env python3
from collections import deque

import fileinput
import sys


def read_grid():
    grid = []
    for line in fileinput.input():
        grid.append(line.strip())
    if len(grid) == 0:
        raise Exception("Empty input")
    width = len(grid[0])
    for i,l in enumerate(grid):
        if len(l) != width:
            raise Exception(f"Line #{i} is not of the expected width {width}.")
    return grid

def solve_grid(grid):
    width = len(grid[0])
    height = len(grid)

    lm_start = None
    pills_pos = []
    for y in range(height):
        for x in range(width):
            if grid[y][x] == 'L':
                if lm_start is not None:
                    raise Exception(f"Multiple Lambdamen found on grid")
                lm_start = (x, y)
            elif grid[y][x] == '.':
                pills_pos.append((x, y))
    if lm_start is None:
        raise Exception("Could not locate Lambdaman")
    print(f"Grid: {width}x{height}, LM@{lm_start}, #Pills: {len(pills_pos)}\n")

    dirs = {'U': (0, -1), 'R': (1, 0), 'D': (0, 1), 'L': (-1, 0)}
    visited = set([lm_start])

    def dfs(curr_pos, pills_to_eat):
        nonlocal min_path
        if pills_to_eat == 0:
            return True
        for move, d in dirs.items():
            new_pos = (curr_pos[0] + d[0], curr_pos[1] + d[1])
            can_go = (0 <= new_pos[0] < width and 0 <= new_pos[1] < height and
                      new_pos not in visited and
                      grid[new_pos[1]][new_pos[0]] != '#')
            if can_go:
                visited.add(new_pos)
                if dfs(new_pos, pills_to_eat - 1):
                    min_path.append(move)
                    return True
                visited.remove(new_pos)
        return False

    min_path = []
    dfs(lm_start, len(pills_pos))
                
    if len(min_path) == 0:
        return "<<NO PATH FOUND>>"
    return "".join(reversed(min_path))


def main():
    grid = read_grid()
    path = solve_grid(grid)
    print(f"\nSolution:\n{path}\n")


if __name__ == "__main__":
    main()
