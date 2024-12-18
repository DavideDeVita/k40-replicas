def DP_probabilities(p, ignore):
    """
    Calculate the probability of at least m true variables over all possible ranges,
    with a full 3D DP table to track starting and ending ranges.

    :param n: Number of variables
    :param p: List of probabilities for each variable
    :return: A 3D DP table dp[i][j][k]
    """
    # Create a 3D DP table initialized to 0
    n = len(p)
    dp = [[[0.0 for _ in range(n + 1)] for _ in range(n)] for _ in range(n)]

    eligibles = dict()

    # Fill DP table for all ranges [i, j]
    for i in range(n):  # Starting node
        for j in range(i, n):  # Ending node
            for k in range(n + 1):  # Exact number of True variables
                if j == i:  # Base case: single node in the range
                    if k == 0:
                        dp[i][j][k] = 1 - p[j]
                    elif k == 1:
                        dp[i][j][k] = p[j]
                else:  # General case: extend the range [i, j-1] to [i, j]
                    dp[i][j][k] = dp[i][j-1][k] * (1 - p[j])  # j-th node is False
                    if k > 0:
                        dp[i][j][k] += dp[i][j-1][k-1] * p[j]  # j-th node is True
                print(i, j, k, dp[i][j][k])

            this_amount = j-i+1
            prob_atleast_half = sum(dp[i][j][(this_amount//2)+1 : ])
            prob_2 = sum(dp[i][j][(this_amount//2)+1 : this_amount+1])
            if prob_2!=prob_atleast_half:
                print("fuck up:", prob_atleast_half, prob_2)
            if prob_atleast_half >= theta:
                if this_amount not in eligibles:
                    eligibles[this_amount] = []
                eligibles[this_amount].append((i, j, prob_atleast_half))

    return dp, eligibles  # Return the full DP table

def DP_getEligibles(p, theta, x_keep=1):
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


# Example usage
if __name__ == "__main__":
    p = [0.9999995, 0.999999, 0.999995, 0.99999, 0.99995, 0.9999, 0.9995, 0.999, 0.995, 0.99] * 50  # Probabilities of each variable being True
    p.sort(reverse = True)
    # p = [0.9999, 0.9995, 0.999, 0.995, 0.99]  # Probabilities of each variable being True
    n = len(p)

    theta = 0.9999999999999

    # Compute the 3D DP table
    dp_table, eligibles_map, pairs, p_skip, triplets = DP_getEligibles(p, theta)

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