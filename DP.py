# Algo 0
def highest_chance_no_issues__ifWorseIgnore(n, probabilities, theta, overkill_size=1):
    # Initialize a 3D DP table
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n + 1)] for _ in range(n + 1)]
    dp[0][0][0] = 1.0  # Base case: No nodes, no selection, no successes
        # k is number of successes

    eligibles = {}
    first_eligible_size=-1

    # Fill the DP table
    for i in range(1, n + 1):  # Iterate over nodes
        for j in range(i + 1):  # Iterate over number of selected nodes // at most i
            # Early stop: If this instance would have "size" greater than min_size + overkill
            if first_eligible_size!=-1 and j > first_eligible_size+overkill_size:
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
            
            #compare with no insert in solution
            prob_if_in_solution = sum(dp[i][j][(j//2)+1:j+1])
            # print(i, j, f"p{(j//2)+1}+:", prob_if_in_solution, "theta:", theta)
            prob_if_not = sum(dp[i-1][j][(j//2)+1:j+1])
            if prob_if_not > prob_if_in_solution:
                # print(i, j, " better excl", prob_if_in_solution, "<", prob_if_not)
                for k in range(j + 1):  # Iterate over number of successes  // at most j
                    dp[i][j][k] = dp[i-1][j][k]
            else:
                #Eligible
                if prob_if_in_solution >= theta:
                    if first_eligible_size==-1 or first_eligible_size>j:
                        # print(f"Found min {i=}, {j=}")
                        first_eligible_size=j

                    if j not in eligibles:
                        eligibles[j] = {}
                    # print("backtracking", i, j)
                    _i = i-1
                    _j = j-1
                    sol = [i]
                    while _j>0:
                        if dp[_i][_j][_j] != dp[_i-1][_j][_j]:
                            sol.append(_i)
                            _j-=1
                        _i-=1
                    eligibles[j][tuple(sol)] = prob_if_in_solution
                    # print("Adding ", tuple(sol), "with p:", prob_if_in_solution)


    for j in range(first_eligible_size+overkill_size+1, n):
        if j in eligibles:
            print(f"{len(eligibles[j])} solutions of size {j} should be removed")
            eligibles.pop(j)

    return dp, eligibles, first_eligible_size+overkill_size
    

# Algo 1 (is like 0 but puts in solution if <theta before chacking if "no put" is better)
def highest_chance_no_issues(n, probabilities, theta, overkill_size=1):
    # Initialize a 3D DP table
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n + 1)] for _ in range(n + 1)]
    dp[0][0][0] = 1.0  # Base case: No nodes, no selection, no successes
        # k is number of successes

    eligibles = {}
    first_eligible_size=-1

    # Fill the DP table
    for i in range(1, n + 1):  # Iterate over nodes
        for j in range(i + 1):  # Iterate over number of selected nodes // at most i
            # Early stop: If this instance would have "size" greater than min_size + overkill
            if first_eligible_size!=-1 and j > first_eligible_size+overkill_size:
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
            
            #compare with no insert in solution
            prob_if_in_solution = sum(dp[i][j][(j//2)+1:j+1])
            # print(i, j, f"p{(j//2)+1}+:", prob_if_in_solution, "theta:", theta)
            prob_if_not = sum(dp[i-1][j][(j//2)+1:j+1])
            if prob_if_in_solution >= theta:
                    # is min size
                    if first_eligible_size==-1 or first_eligible_size>j:
                        # print(f"Found min {i=}, {j=}")
                        first_eligible_size=j

                    if j not in eligibles:
                        eligibles[j] = {}
                    # print("backtracking", i, j)
                    _i = i-1
                    _j = j-1
                    sol = [i]
                    while _j>0:
                        if dp[_i][_j][_j] != dp[_i-1][_j][_j]:
                            sol.append(_i)
                            _j-=1
                        _i-=1
                    eligibles[j][tuple(sol)] = prob_if_in_solution
                    # print("Adding ", tuple(sol), "with p:", prob_if_in_solution)

            if prob_if_not > prob_if_in_solution:
                # print(i, j, " better excl", prob_if_in_solution, "<", prob_if_not)
                for k in range(j + 1):  # Iterate over number of successes  // at most j
                    dp[i][j][k] = dp[i-1][j][k]
                


    for j in range(first_eligible_size+overkill_size+1, n):
        if j in eligibles:
            print(f"{len(eligibles[j])} solutions of size {j} should be removed")
            eligibles.pop(j)

    return dp, eligibles, first_eligible_size+overkill_size


# Algo 2 (is like 0 but puts in solution if <theta before chacking if "no put" is better)
def highest_chance_no_issues__Break(n, probabilities, theta):
    # Initialize a 3D DP table
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n + 1)] for _ in range(n + 1)]
    dp[0][0][0] = 1.0  # Base case: No nodes, no selection, no successes
        # k is number of successes

    first_eligible, first_eligible_size, el_prob = (), -1, 1.

    # Fill the DP table
    for i in range(1, n + 1):  # Iterate over nodes
        for j in range(i + 1):  # Iterate over number of selected nodes // at most i
            if first_eligible_size>0 and j>first_eligible_size:
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
            
            if first_eligible_size<1 or j<first_eligible_size:
                #compare with no insert in solution
                prob_if_in_solution = sum(dp[i][j][(j//2)+1:j+1])
                print(i, j, f"p{(j//2)+1}+:", prob_if_in_solution, "theta:", theta)
                if prob_if_in_solution >= theta:
                        # print("backtracking", i, j)
                        _i = i-1
                        _j = j-1
                        sol = [i]
                        while _j>0:
                            if dp[_i][_j][_j] != dp[_i-1][_j][_j]:
                                sol.append(_i)
                                _j-=1
                            _i-=1
                        first_eligible, first_eligible_size, el_prob = tuple(sol), len(sol), prob_if_in_solution
                        print("---------- Adding ", tuple(sol), "with p:", prob_if_in_solution)

    return dp, first_eligible, el_prob


def permutate_min_solution (probabilities, firstSolution, theta):
    pass

# Algo 3
def DP_getEligibles_streak(p, theta, overkill_size=1):
    """
    Calculate the probability of at least m true variables over all possible ranges,
    with a full 3D DP table to track starting and ending ranges.

    :param theta: Lowest acceptable probability
    :param p: List of probabilities for each variable
    :param overkill_size: When it finds the first eligible group, skip any solution that has length (size) greater than size+overkill_size
    :return: A 3D DP table dp[i][j][k]
    :return: A list of tuples (size, start, end, prob_half+) of eligible intervals //size is redundant but who cares
    """
    # Create a 3D DP table initialized to 0
    n = len(p)
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n)] for _ in range(n)]
    first_el_size = -1
    eligibles = dict()
    triplets_count = 0
    pairs_count = 0
    pairs_skipped = 0

    # Fill DP table for all ranges [i, j]
    for i in range(n):  # Starting node
        for j in range(i, n):  # Ending node
            this_size = (j - i) + 1
            # Skip condition
            if first_el_size != -1 and this_size > first_el_size+overkill_size:
                pairs_skipped+=1
                continue

            pairs_count +=1

            for k in range(this_size + 1):  # Exact number of True variables
                triplets_count+=1
                if j == i:  # Base case: single node in the range
                    if k == 0:
                        dp[i][j][k] = 1 - p[j]
                    elif k == 1:
                        dp[i][j][k] = p[j]
                else:  # General case: extend the range [i, j-1] to [i, j]
                    dp[i][j][k] = dp[i][j-1][k] * (1 - p[j])  # j-th node is False
                    if k > 0:
                        dp[i][j][k] += dp[i][j-1][k-1] * p[j]  # j-th node is True
            
            prob_atleast_half = sum(dp[i][j][(this_size//2)+1 : ])
            if prob_atleast_half >= theta:
                if overkill_size>0 and first_el_size==-1:
                    first_el_size = this_size

                if this_size not in eligibles:
                    eligibles[this_size] = []
                eligibles[this_size].append((i, j, prob_atleast_half))


    return dp, eligibles, pairs_count, pairs_skipped, triplets_count  # Return the full DP table


# Algo 4
from itertools import combinations
def search_all_combinations(p, theta, *, overkill_size=1):
    nodes = [i for i in range(len(p))]
    eligibles = {}
    first_eligible_size = -1

    for size in range(1, len(p)):
        n = size
        if first_eligible_size!=-1 and n>first_eligible_size+overkill_size:
            break
        for combo in combinations(nodes, n):
            probs = [p[i] for i in combo]
            # DP table where dp[i][s] stores probability of exactly s successes among first i nodes
            dp = [[0.0] * (n + 1) for _ in range(n + 1)]
            
            # Base case: Probability of 0 successes in 0 nodes is 1
            dp[0][0] = 1.0

            # Fill DP table
            for i in range(1, n + 1):  # Iterate over nodes
                for s in range(i + 1):  # Iterate over possible success counts
                    dp[i][s] = dp[i - 1][s] * (1 - probs[i - 1])  # Node i fails
                    if s > 0:
                        dp[i][s] += dp[i - 1][s - 1] * probs[i - 1]  # Node i succeeds

            # Sum probabilities for at least ceil(n/2) successes
            k = n//2 + 1
            p_atLeast_half = sum(dp[n][k:])
            if p_atLeast_half>=theta:
                # print(f"{combo}:", end="\t")
                # for ii in range(len(dp[n])):
                #     print(f"{ii}:{dp[n][ii]:.7f}", end="\t")
                # print(f"p({k}+) = {p_atLeast_half:.20} % > {theta}")
                if first_eligible_size==-1:
                    first_eligible_size = size

            
                if size not in eligibles:
                    eligibles[size] = []
                eligibles[size].append( combo + (p_atLeast_half,) )
    return eligibles



def is_eligible(p, sol, theta):
    s = sum( [p[i-1] for i in sol])
    return s>=theta*len(sol)    


# TO DO:
#     > In insert_pod_dp try aorting per score
#     > Here, try new algo stopping when you find the first eligible (you can propagate from there e.g. if (0 3 5) is first eligible any x y z where x>=0, y>=3, z>=5 )

import time
# Example usage
algo=2
if __name__ == "__main__":
    p = [0.9999995, 0.999999, 0.999995, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99] * 5  # Probabilities of each variable being True
    # p = [0.999999, 0.999995, 0.99999, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True
    n = len(p)

    theta = 0.999999995
    p.sort()


    if algo==0:
        # Example Usage
        dp, eligibles, maxSize = highest_chance_no_issues__ifWorseIgnore(n, p, theta) #, overkill_size=2)

        print("\n\nAlgo", algo)

        if n<15:
            for i in range(1, n+1):
                print(f"Accounting {i} nodes.\t\tP[{i-1}]={p[i-1]}")
                for j in range(min(i, maxSize)+1):
                    print(f"\t{j} replicas:-", end="\t")
                    for k in range(j+1):   #from 0 to distance(i, j) (included)
                        print(f"{k}: {dp[i][j][k]:.10f}", end="   ")

                    prob_atleast_half = sum(dp[i][j][(j//2)+1 : ])
                    s = ""
                    if prob_atleast_half >= theta:
                        s = " > θ"
                    print((("     "+(" "*10)+"   ")*(maxSize-j)) + f"P[{(j//2) +1} +]: {prob_atleast_half:.10f}"+s)
                    print()
                print()
        
        for size in range(1, maxSize+1):
            if size in eligibles:
                print(f"Using {size} nodes, {len(eligibles[size])} eligible solutions:")
                for t_sol, prob_h in eligibles[size].items():
                    print(f"\t{t_sol}: {[p[i-1] for i in t_sol]} -> {prob_h}")


    elif algo==1:
        # Example Usage
        dp, eligibles, maxSize = highest_chance_no_issues(n, p, theta)

        print("\nAlgo", algo)

        if n<15:
            for i in range(1, n+1):
                print(f"Accounting {i} nodes.\t\tP[{i-1}]={p[i-1]}")
                for j in range(min(i, maxSize)+1):
                    print(f"\t{j} replicas:-", end="\t")
                    for k in range(j+1):   #from 0 to distance(i, j) (included)
                        print(f"{k}: {dp[i][j][k]:.10f}", end="   ")

                    prob_atleast_half = sum(dp[i][j][(j//2)+1 : ])
                    s = ""
                    if prob_atleast_half >= theta:
                        s = " > θ"
                    print((("     "+(" "*10)+"   ")*(maxSize-j)) + f"P[{(j//2) +1} +]: {prob_atleast_half:.10f}"+s)
                    print()
                print()
        
        for size in range(1, maxSize+1):
            if size in eligibles:
                print(f"Using {size} nodes, {len(eligibles[size])} eligible solutions:")
                for t_sol, prob_h in eligibles[size].items():
                    print(f"\t{t_sol}: {[p[i-1] for i in t_sol]} -> {prob_h}")


    elif algo==2:
        def all_unique(lst):
            return len(lst) == len(set(lst))
        
        # p.sort()

        # Example Usage
        dp, firstEligible, firstEligibleProb = highest_chance_no_issues__Break(n, p, theta)

        print("\nAlgo", algo)

        print(firstEligible, "prob:", firstEligibleProb)

        eligibles_neigh = [firstEligible]
        size = len(firstEligible)

        # permutate
        i = 0
        while i<len(eligibles_neigh):
            el = eligibles_neigh[i]
            print(i, ":", el)
            for idx in range(size):
                for jdx in range(size):
                    if idx==jdx:
                        continue
                    if el[idx] == n or el[jdx]==1:
                        continue
                    newT = list(el)
                    newT[idx] += 1
                    newT[jdx] -= 1
                    newT.sort(reverse=True)
                    newT = tuple(newT)
                    print("newtT=", newT)
                    if all_unique(newT) and newT not in eligibles_neigh:
                        if  is_eligible(p, newT, theta):
                            print("appending", newT)
                            eligibles_neigh.append(tuple(newT))
            i+=1

        # all greater
        eligibles = set(eligibles_neigh)

        for el in eligibles_neigh:
            newT = list(el[:])
            print(el)
            while newT[0] <= n:
                newT[-1] += 1
                if newT[-1] > n:
                    idx = size-1
                    while idx>0 and newT[idx] > n:
                        newT[idx] = el[idx]
                        newT[idx-1] += 1
                        idx -= 1
                
                if all_unique(newT):
                    eligibles.add(tuple(newT))
        
        print(f"Using {size} nodes, {len(eligibles)} eligible solutions:")
        for t_sol in eligibles:
            print(f"\t{t_sol}: ")
            print(f"\t{t_sol}: {[p[i-1] for i in t_sol]}")

        
    elif algo==3:
        p.sort(reverse = True)
        # p = [0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True


        # Compute the 3D DP table
        dp_table, eligibles_map, pairs, p_skip, triplets = DP_getEligibles_streak(p, theta)

        # Print results for any range
        if n<15:
            for i in range(n):
                print(f"Starting from node {i} (prob: {p[i]})")
                for j in range(i, n):
                    size = j-i+1
                    print(f"\t to {j} ({size} repl):- ", end="\t")
                    for k in range(size +1):   #from 0 to distance(i, j) (included)
                        print(f"{k}: {dp_table[i][j][k]:.7f}", end="    ")
                        
                    prob_atleast_half = sum(dp_table[i][j][(size//2)+1 : ])
                    print((("     "+(" "*7)+"    ")*(n-size)) + f"P[{(size//2) +1} +]: {prob_atleast_half:.7f}")

                print("\n")

        print(f"theta:\t         {theta:.12f}")
        last_size = -1
        for size, eligibles in eligibles_map.nodes():
            print(f"{size} replicas, {len(eligibles)} eligibles considered: ")
            for x in eligibles:
                print(f"\t[{x[0]} - {x[1]}]: {x[2]:.20f}")
            print()

        #complexity
        print(f"Ran {pairs} pairs out of exp {n**2}: {(100.*pairs/n**2):.4f}%\nRan {pairs} pairs out of really {p_skip+pairs}: {(100.*pairs/(p_skip+pairs)):.4f}%\nSkipped {p_skip} pairs out of {pairs+p_skip}: {(100.*p_skip/(pairs+p_skip)):.4f}%\nTriplets ran: {triplets} out of {n**3}: {(100.*triplets/n**3):.4f}%\n")


    elif algo == 4:
        # Example Usage
        eligibles = search_all_combinations(p, theta, overkill_size=0)
        

        for size in range(1, n+1):
            if size in eligibles:
                print("Using ", size, "nodes. ", len(eligibles[size])," available solutions\n")

                for sol in eligibles[size]:
                    print(f"\tNodes: {sol[:-1]}\t p: {sol[-1]:.12f} %\n")



if __name__ == "__main__for":
    p = [0.9999995, 0.999999, 0.999995, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99] * 5  # Probabilities of each variable being True
    # p = [0.999999, 0.999995, 0.99999, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True
    n = len(p)
    t = []
    sol = [0 for _ in range(4)]

    theta = 0.99999995

    p.sort()

    for algo in range(3):   #(4)

        start = time.perf_counter()


        if algo==0:
            # Example Usage
            dp, eligibles, maxSize = highest_chance_no_issues__ifWorseIgnore(n, p, theta) #, overkill_size=2)

            print("\n\nAlgo", algo)

            if n<15:
                for i in range(1, n+1):
                    print(f"Accounting {i} nodes.\t\tP[{i-1}]={p[i-1]}")
                    for j in range(min(i, maxSize)+1):
                        print(f"\t{j} replicas:-", end="\t")
                        for k in range(j+1):   #from 0 to distance(i, j) (included)
                            print(f"{k}: {dp[i][j][k]:.10f}", end="   ")

                        prob_atleast_half = sum(dp[i][j][(j//2)+1 : ])
                        s = ""
                        if prob_atleast_half >= theta:
                            s = " > θ"
                        print((("     "+(" "*10)+"   ")*(maxSize-j)) + f"P[{(j//2) +1} +]: {prob_atleast_half:.10f}"+s)
                        print()
                    print()
            
            for size in range(1, maxSize+1):
                if size in eligibles:
                    print(f"Using {size} nodes, {len(eligibles[size])} eligible solutions:")
                    for t_sol, prob_h in eligibles[size].items():
                        print(f"\t{t_sol}: {[p[i-1] for i in t_sol]} -> {prob_h}")

                    sol[algo] += len(eligibles[size])


        elif algo==1:
            # Example Usage
            dp, eligibles, maxSize = highest_chance_no_issues(n, p, theta)

            print("\nAlgo", algo)

            if n<15:
                for i in range(1, n+1):
                    print(f"Accounting {i} nodes.\t\tP[{i-1}]={p[i-1]}")
                    for j in range(min(i, maxSize)+1):
                        print(f"\t{j} replicas:-", end="\t")
                        for k in range(j+1):   #from 0 to distance(i, j) (included)
                            print(f"{k}: {dp[i][j][k]:.10f}", end="   ")

                        prob_atleast_half = sum(dp[i][j][(j//2)+1 : ])
                        s = ""
                        if prob_atleast_half >= theta:
                            s = " > θ"
                        print((("     "+(" "*10)+"   ")*(maxSize-j)) + f"P[{(j//2) +1} +]: {prob_atleast_half:.10f}"+s)
                        print()
                    print()
            
            for size in range(1, maxSize+1):
                if size in eligibles:
                    print(f"Using {size} nodes, {len(eligibles[size])} eligible solutions:")
                    for t_sol, prob_h in eligibles[size].items():
                        print(f"\t{t_sol}: {[p[i-1] for i in t_sol]} -> {prob_h}")
                    
                    sol[algo] += len(eligibles[size])


        elif algo==2:
            def all_unique(lst):
                return len(lst) == len(set(lst))
            # p.sort()

            # Example Usage
            dp, firstEligible, firstEligibleProb = highest_chance_no_issues__Break(n, p, theta)

            print("\nAlgo", algo)

            eligibles_neigh = [firstEligible]
            size = len(firstEligible)

            # permutate
            i = 0
            while i<len(eligibles_neigh):
                el = eligibles_neigh[i]
                print(i, ":", el)
                for idx in range(size):
                    for jdx in range(size):
                        if idx==jdx:
                            continue
                        newT = list(el)
                        newT[idx] += 1
                        newT[jdx] -= 1
                        newT.sort(reverse=True)
                        newT = tuple(newT)
                        print("newtT=", newT)
                        if all_unique(newT) and newT not in eligibles_neigh:
                            if  is_eligible(p, newT, theta):
                                print("appending", newT)
                                eligibles_neigh.append(tuple(newT))
                i+=1

            # all greater
            eligibles = set(eligibles_neigh)

            for el in eligibles_neigh:
                newT = list(el[:])
                while newT[0] <= n:
                    newT[-1] += 1
                    if newT[-1] > n:
                        idx = size-1
                        while idx>0 and newT[idx] > n:
                            newT[idx] = el[idx]
                            newT[idx-1] += 1
                            idx -= 1
                    
                    if all_unique(newT):
                        eligibles.add(tuple(newT))
            
            print(f"Using {size} nodes, {len(eligibles[size])} eligible solutions:")
            for t_sol in eligibles:
                print(f"\t{t_sol}: {[p[i-1] for i in t_sol]}")
            
            sol[algo] = len(eligibles)

        

        elif algo==3:
            p.sort(reverse = True)
            # p = [0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True


            # Compute the 3D DP table
            dp_table, eligibles_map, pairs, p_skip, triplets = DP_getEligibles_streak(p, theta)

            # Print results for any range
            if n<15:
                for i in range(n):
                    print(f"Starting from node {i} (prob: {p[i]})")
                    for j in range(i, n):
                        size = j-i+1
                        print(f"\t to {j} ({size} repl):- ", end="\t")
                        for k in range(size +1):   #from 0 to distance(i, j) (included)
                            print(f"{k}: {dp_table[i][j][k]:.7f}", end="    ")
                            
                        prob_atleast_half = sum(dp_table[i][j][(size//2)+1 : ])
                        print((("     "+(" "*7)+"    ")*(n-size)) + f"P[{(size//2) +1} +]: {prob_atleast_half:.7f}")

                    print("\n")

            print(f"theta:\t         {theta:.12f}")
            last_size = -1
            for size, eligibles in eligibles_map.nodes():
                print(f"{size} replicas, {len(eligibles)} eligibles considered: ")
                sol[algo-1] += len(eligibles)

                for x in eligibles:
                    print(f"\t[{x[0]} - {x[1]}]: {x[2]:.20f}")
                print()

            #complexity
            print(f"Ran {pairs} pairs out of exp {n**2}: {(100.*pairs/n**2):.4f}%\nRan {pairs} pairs out of really {p_skip+pairs}: {(100.*pairs/(p_skip+pairs)):.4f}%\nSkipped {p_skip} pairs out of {pairs+p_skip}: {(100.*p_skip/(pairs+p_skip)):.4f}%\nTriplets ran: {triplets} out of {n**3}: {(100.*triplets/n**3):.4f}%\n")


        elif algo == 4:
            # Example Usage
            eligibles = search_all_combinations(p, theta)
            

            for size in range(1, n+1):
                if size in eligibles:
                    print("Using ", size, "nodes. ", len(eligibles[size])," available solutions\n")
                    sol[algo-1] += len(eligibles[size])

                    for s in eligibles[size]:
                        print(f"\tNodes: {s[:-1]}\t p: {s[-1]:.12f} %\n")


        # Your code snippet here
        end = time.perf_counter()
        t.append(end-start)

    for algo in range(3):   #4
        print(f"Algo {algo}: {sol[algo]} solutions\t\t{t[algo]:.12f} seconds")