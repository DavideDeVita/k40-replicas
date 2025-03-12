def find_eligible_solution(n, probabilities, theta, overkill_size=0):
    # Initialize a 3D DP table
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n + 1)] for _ in range(n + 1)]
    dp[0][0][0] = 1.0  # Base case: No nodes, no selection, no successes
        # k is number of successes

    first_eligible, first_eligible_size, first_eligible_prob = {}, -1, {}

    # Fill the DP table
    for i in range(1, n + 1):  # Iterate over nodes
        for j in range(i + 1):  # Iterate over number of selected nodes // at most i
            if first_eligible_size>0 and j>first_eligible_size+overkill_size:
                continue

            #compute the "if (i-1) added to [i-1][j-1] line". If sum from j//2+1 > sum for [i-1][j] save line, else repeat previous
            for k in range(j + 1):  # Iterate over number of successes  // at most j
                # default is always 0
                # Include the i-th item
                if j > 0:
                    dp[i][j][k] += dp[i-1][j-1][k] * (1. -probabilities[i-1])   # In solution but fails
                    if k > 0: #Can account for a success?
                        dp[i][j][k] += dp[i-1][j-1][k-1] * probabilities[i-1]   # In solution but success
                else:
                    dp[i][j][k] = 1.
                # print(f"dp[{i}][{j}][{k}] = {dp[i][j][k]}")
            
            ###
            # Eval solution
            if j>0:
                if first_eligible_size<1 or j<first_eligible_size:
                    prob_if_in_solution = sum(dp[i][j][(j//2)+1:j+1])
                    # print(i, j, f"p{(j//2)+1}+:", prob_if_in_solution, "\t\ttheta: ", theta)
                    if prob_if_in_solution >= theta:        # is eligible?
                            # print("backtracking", i, j)
                            _i = i-1
                            _j = j-1
                            sol = [i]
                            while _j>0:
                                if dp[_i][_j][_j] != dp[_i-1][_j][_j]:
                                    sol.append(_i-1)        # 'i' index is shifted by one
                                    _j-=1
                                _i-=1
                            sol = tuple(sol)
                            first_eligible[j], first_eligible_size, first_eligible_prob[j] = sol, j, prob_if_in_solution
                            print("---------- Adding ", sol, "with p:", prob_if_in_solution)
    
    for size in range(first_eligible_size+overkill_size+1, n):
        # print("removing size", size)
        first_eligible.pop(size, None)
        first_eligible_prob.pop(size, None)

    return dp, first_eligible, first_eligible_prob


def permutate_min_solution (probabilities, firstEligible, theta, neigh_search=1):
    eligibles_neigh = [firstEligible]
    size = len(firstEligible)

    # Searching for each eligible (wide search)
    i = 0
    while i<len(eligibles_neigh):
        _sol = eligibles_neigh[i]
        # print(i, ":", _sol)
        for idx in range(size):     # I increment idx 
            for jdx in range(size):     # I decrement jdx
                if idx==jdx:
                    continue

                _sol_L = list(_sol)
                for shift in range(neigh_search+1): #(-neigh_search, neigh_search+1):
                    if shift==0:
                        continue
                    neigh = _sol_L[:]
                    neigh[idx] += shift
                    neigh[jdx] -= shift
                    # # If either is overflow or
                    # #     if incremented reaches the following node index (Remember: indexes in these tuples are sorted decr [es: (10, 8, 7)]; if decrementing a val in idx i reach the val in idx+1 means i have repetition) 
                    # #         same but for jdx
                    if neigh[idx]>=n or neigh[jdx]<0 or \
                        (idx>0 and neigh[idx]>=_sol_L[idx-1]) or \
                            (jdx<size-1 and neigh[jdx]<=_sol_L[jdx+1]) or \
                                (idx>0 and neigh[idx]>=neigh[idx-1]) or \
                                    (jdx<size-1 and neigh[jdx]<=neigh[jdx+1]):
                        # print(f"Skipping: {idx=}, {jdx=},\n\t {neigh[idx]=}, {neigh[jdx]=},\n\t{neigh=}, {_sol_L=}, {shift=}")
                        continue
                    neigh = tuple(neigh)
                    # print("neigh=", neigh)
                    if neigh not in eligibles_neigh:    # all_unique(neigh) and 
                        if  is_eligible(probabilities, neigh, theta):
                            # print("appending", neigh)
                            eligibles_neigh.append(neigh)
        i+=1
    return eligibles_neigh


def is_eligible(p, sol, theta):
    p=1
    for x in sol:
        p*=x
    return p>=20000    

def all_unique(lst):
    if len(lst) != len(set(lst)):
        print("UNEXPECTED", lst)
    return len(lst) == len(set(lst))

def all_greater_tuples(input_eligibles, n):
    # all greater
    eligibles = set(input_eligibles)

    for el in input_eligibles:
        size = len(el)
        # print(el)
        newT = list(el[:])
        while newT[0] < n:
            newT[-1] += 1
            if newT[-1] >= newT[-2]:
                idx = size-1
                while idx>0 and newT[idx] >= newT[idx-1]:
                    newT[idx] = el[idx]
                    newT[idx-1] += 1
                    idx -= 1
            
            if newT[0] < n:
                eligibles.add(tuple(newT))
    return eligibles

def custom_sort(tuples_set):
    return sorted(tuples_set, key=lambda x: (x, len(x)))

# TO DO:
#     > In insert_pod_dp try aorting per score
#     > Here, try new algo stopping when you find the first eligible (you can propagate from there e.g. if (0 3 5) is first eligible any x y z where x>=0, y>=3, z>=5 )

import time
# Example usage
if __name__ == "__main__":
    p = [0.9999995, 0.999999, 0.999995, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99] * 2  # Probabilities of each variable being True
    # p = [0.999999, 0.999995, 0.99999, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True
    p.append(0.9999998)
    n = len(p)

    theta = 0.999999995
    p.sort()

    # Example Usage
    dp, first_eligibles, first_eligibles_prob = find_eligible_solution(n, p, theta, overkill_size=2)

    all_eligibles = set()

    for size, el_sol in first_eligibles.items():
        print(f"Solution of size {size}: {el_sol}")

        eligibles_neigh = permutate_min_solution(p, el_sol, theta, neigh_search=2)
    
        print(f"{len(eligibles_neigh)} neigh solutions of size {size}:")
        # for neigh in eligibles_neigh:
        #     print(f"\t{neigh}: {[p[i] for i in neigh]}")

        all_eligibles.update(all_greater_tuples(eligibles_neigh, n))


    all_eligibles = custom_sort(all_eligibles)
    print(f"{len(all_eligibles)} neigh solutions:")
    i=2
    for sol in all_eligibles:
        if i>0:
            print(f"\t{sol}: {[p[i] for i in sol]}", end="\t")
            i-=1
        else:
            print(f"\t{sol}: {[p[i] for i in sol]}")
            i=2
