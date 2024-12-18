import numpy as np
from typing import Dict, List, Set

p = [ 0.1, 0.25, 0.2, 0.1, 0.1, 0.15, 0.3, 0.1, 0.1, 0.05]  # P of fail
p.sort(reverse=True)
print(p)
N = len(p)
R = N

# Initialize the Choice array to track decisions
# Choice = [[[False for k in range(R+1)] for j in range(R+1)] for i in range(N+1)]
Choice = np.zeros((N + 1, R + 1, R + 1), dtype=bool)


# Re-initialize the DP table
DP = np.zeros((N + 1, R + 1, R + 1))
DP[0][0][0] = 1.0  # Base case

# Fill the DP table with choice tracking
for k in range(R + 1):
    for i in range(1, N + 1):
        for j in range(R + 1):

            # If k=0, we search for the Probability of no failures when pickng j out of i nodes.
            # We therefore want to maximize that probability, to find the best subset
            if k==0:
                if j>i:
                   # Need more nodes than I have in total. Keep P to zero and Choice to zero
                   continue

                if j==0:
                    # I have nodes but I don't need to pick anything, default as a No-Pick
                    DP[i][j][k] = DP[i-1][j][k]
                    ## Implicit: Choice[i][j][k] = False  # Not selecting item i-1
                    continue

                # Inserting i-th in solution is better (greater chance) than not having it?
                no_pick = DP[i-1][j][k]
                pick = DP[i-1][j-1][k] * (1-p[i-1])
                if  pick >= no_pick: 
                    DP[i][j][k] = pick
                    Choice[i][j][k] = True  # Item selected
                else:
                    DP[i][j][k] = no_pick
                    
            else:
                if j>i or k>j:
                   # Limit impossible case
                   continue

                # Prob of ... on pick is the sum of prob of istance with one less fail when this fails, plus same amount of fails when this doesn't fail
                no_pick = DP[i-1][j][k]
                pick = DP[i-1][j-1][k] * (1-p[i-1])  +  DP[i-1][j-1][k-1] * (p[i-1])
                if no_pick==0 or no_pick >= pick:
                    DP[i][j][k] = pick
                    Choice[i][j][k] = True
                else:
                    DP[i][j][k] = no_pick



# Print the DP table to observe its structure
for k in range(R + 1):
    print(f"{k=}:")
    s = ""
    for i in range(N + 1):
        if k>i:
            continue
        s = f"{i}:"
        for j in range(R + 1):
            if j>i or k>j:
                # Limit impossible case
                continue
            c = ""
            if Choice[i][j][k]:
                c = "*"
            s += f"\t{c}[{j}] = {DP[i][j][k]:.5f}"
            # s += f"\t[{j}] = {Choice[i][j][k]}"
        print(s)
    print("\n")  # Separate levels for better readability


theta_p = 0.125
# Compute the aggr matrix: Summing the ij matrixes for k>=(R+1)//2
    # This will be a matrix with prob that at least half fail, we then want to sort them by j(number of nodes deployed) 

# i, j, k = N, R, (R+1)//2
prob_atLeastHalfFail = np.zeros((N+1, R+1))
for i in range(1, N+1):
    for j in range(1, R + 1):
        for k in range( (j+1)//2, j+1 ):
            prob_atLeastHalfFail[i][j] += DP[i][j][k]

print("Prob of Half+")
eligible = [[] for j in range(R+1)]
for i in range(1, N + 1):
    s = f"{i}:"
    for j in range(1, R + 1):
        if j>i:
            break
        gud = ""
        if Choice[i][j][0] and prob_atLeastHalfFail[i][j] < theta_p:
            gud = "*"
            eligible[j].append(i)
        s += f"\t{gud}[{j}] = {prob_atLeastHalfFail[i][j]:.3f}"
    print(s)

print()

# Backtrack to find the solutions
solutions:Dict[int,List[List[int]]] = {}

#PRINT
for j in range(R+1):
    if len(eligible[j])>0:
        solutions[j] = []
        print(f"Found at least {len(eligible[j])} eligible solutions with {j} nodes: {eligible[j]}")
        
        for i in eligible[j]:
            selected_items = []
            k = j
            _j = j
            _i = i
            while _j > 0:
                # print(f"Choice[{i}][{_j}][{k}]:{Choice[i][_j][k]}")
                if Choice[_i][_j][k]:
                    selected_items.append(_i)  # Add item index (0-based)
                    _j -= 1  # Move to previous state
                    k -= 1  # Move to previous state
                _i -= 1  # Move to the previous item

            selected_items.reverse()  # Reverse to get the correct order
            f = f"{selected_items}"
            f += f": {prob_atLeastHalfFail[i][j]:.3f}"
            print(f)
            solutions[j].append(selected_items)



print()
print(f"{theta_p=}")
for i in range(len(p)):
    print(f"{i+1}:{p[i]}", end="   ")
print("\n")

for k,v in solutions.items():
    print(k, ":\t", v)
