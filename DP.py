def highest_chance_no_issues(n, probabilities):
    # Initialize a 3D DP table
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n + 1)] for _ in range(n + 1)]
    dp[0][0][0] = 1.0  # Base case: No items, no selection, no successes
        # k is number of successes

    prob_atleast_half = [0 for j in range(n+1)]

    # Fill the DP table
    for i in range(1, n + 1):  # Iterate over items
        for j in range(i + 1):  # Iterate over number of selected items // at most i
            #compute the "if (i-1) added to [i-1][j-1] line". If sum from j//2+1 > sum for [i-1][j] save line, else repeat previous
            for k in range(j + 1):  # Iterate over number of successes  // at most j
                # default is always 0
                # Include the i-th item
                if j > 0:
                    dp[i][j][k] += dp[i-1][j-1][k] * (1 - probabilities[i-1])   # In solution but fails
                    if k > 0: #Can account for a success?
                        dp[i][j][k] += dp[i-1][j-1][k-1] * probabilities[i-1]   # In solution but success
                else:
                    dp[i][j][k] = 1.
            
            #compare with no insert in solution
            prob_if_in_solution = sum(dp[i][j][(j//2)+1:j+1])
            prob_if_not = sum(dp[i-1][j][(j//2)+1:j+1])
            if prob_if_not > prob_if_in_solution:
                print(i, j, " better excl", prob_if_in_solution, "<", prob_if_not)
                for k in range(j + 1):  # Iterate over number of successes  // at most j
                    dp[i][j][k] = dp[i-1][j][k]
                
                if i == n:
                    prob_atleast_half[j] = prob_if_not
            else:
                if i == n:
                    prob_atleast_half[j] = prob_if_in_solution

    return dp, prob_atleast_half



def DP_getEligibles_streak(p, theta, x_keep=1):
    """
    Calculate the probability of at least m true variables over all possible ranges,
    with a full 3D DP table to track starting and ending ranges.

    :param theta: Lowest acceptable probability
    :param p: List of probabilities for each variable
    :param x_keep: When it finds the first eligible group, skip any solution that has length (amount) greater than amount+x_keep
    :return: A 3D DP table dp[i][j][k]
    :return: A list of tuples (amount, start, end, prob_half+) of eligible intervals //amount is redundant but who cares
    """
    # Create a 3D DP table initialized to 0
    n = len(p)
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n)] for _ in range(n)]
    first_el_amount = -1
    eligibles = dict()
    triplets_count = 0
    pairs_count = 0
    pairs_skipped = 0

    # Fill DP table for all ranges [i, j]
    for i in range(n):  # Starting node
        for j in range(i, n):  # Ending node
            this_amount = (j - i) + 1
            # Skip condition
            if first_el_amount != -1 and this_amount > first_el_amount+x_keep:
                pairs_skipped+=1
                continue

            pairs_count +=1

            for k in range(this_amount + 1):  # Exact number of True variables
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
            
            prob_atleast_half = sum(dp[i][j][(this_amount//2)+1 : ])
            if prob_atleast_half >= theta:
                if x_keep>0 and first_el_amount==-1:
                    first_el_amount = this_amount

                if this_amount not in eligibles:
                    eligibles[this_amount] = []
                eligibles[this_amount].append((i, j, prob_atleast_half))


    return dp, eligibles, pairs_count, pairs_skipped, triplets_count  # Return the full DP table



TO DO:
    > In insert_pod_dp try aorting per score
    > Here, try new algo stopping when you find the first eligible (you can propagate from there e.g. if (0 3 5) is first eligible any x y z where x>=0, y>=3, z>=5 )

algo=2
# Example usage
if __name__ == "__main__":
    # p = [0.9999995, 0.999999, 0.999995, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99] * 50  # Probabilities of each variable being True
    p = [0.999999, 0.999995, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True
    p.sort()
    n = len(p)
    
    theta = 0.99

    if algo==1:
        # Example Usage
        dp, probs_per_j = highest_chance_no_issues(n, p)

        if n<15:
            for i in range(1, n+1):
                print(f"Accounting {i} nodes.\t\tP[{i-1}]={p[i-1]}")
                for j in range(i+1):
                    print(f"\t{j} replicas:- ", end="\t")
                    for k in range(j+1):   #from 0 to distance(i, j) (included)
                        print(f"{k}: {dp[i][j][k]:.15f}", end="    ")
                    print()
                print()


        for j in range(1, len(probs_per_j)):
            s = ""
            if probs_per_j[j]>=theta:
                s = " ≥ θ"
            print(f"{j} replicas - Prob[{1+(j//2)}+] = {probs_per_j[j]:.15f} {s}\n")

    elif algo==2:
        p.sort(reverse = True)
        # p = [0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True


        # Compute the 3D DP table
        dp_table, eligibles_map, pairs, p_skip, triplets = DP_getEligibles_streak(p, theta)

        # Print results for any range
        if n<15:
            for i in range(n):
                print(f"Starting from node {i} (prob: {p[i]})")
                for j in range(i, n):
                    amount = j-i+1
                    print(f"\t to {j} ({amount} repl):- ", end="\t")
                    for k in range(amount +1):   #from 0 to distance(i, j) (included)
                        print(f"{k}: {dp_table[i][j][k]:.7f}", end="    ")
                        
                    prob_atleast_half = sum(dp_table[i][j][(amount//2)+1 : ])
                    print((("     "+(" "*7)+"    ")*(n-amount)) + f"P[{(amount//2) +1} +]: {prob_atleast_half:.7f}")

                print("\n")

        print(f"theta:\t         {theta:.12f}")
        last_amount = -1
        for amount, eligibles in eligibles_map.items():
            print(f"{amount} replicas, {len(eligibles)} eligibles considered: ")
            for x in eligibles:
                print(f"\t[{x[0]} - {x[1]}]: {x[2]:.20f}")
            print()

        #complexity
        print(f"Ran {pairs} pairs out of exp {n**2}: {(100.*pairs/n**2):.4f}%\nRan {pairs} pairs out of really {p_skip+pairs}: {(100.*pairs/(p_skip+pairs)):.4f}%\nSkipped {p_skip} pairs out of {pairs+p_skip}: {(100.*p_skip/(pairs+p_skip)):.4f}%\nTriplets ran: {triplets} out of {n**3}: {(100.*triplets/n**3):.4f}%\n")