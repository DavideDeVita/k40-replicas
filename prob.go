package main

func compute_probability_atLeastHalf(p []float64) float64 {
	n := len(p)
	// Initialize dp array with size n+1
	dp := make([]float64, n+1)
	dp[0] = 1.0 // Base case: probability of 0 true variables is 1

	// Iterate over each probabilistic variable
	for i := 0; i < n; i++ {
		// Update dp array in reverse (to avoid overwriting values we still need)
		for k := n; k > 0; k-- {
			dp[k] = dp[k]*(1-p[i]) + dp[k-1]*p[i]
		}
		// Update dp[0] (probability of no true variables)
		dp[0] = dp[0] * (1 - p[i])
	}

	// Sum up probabilities for at least m true variables
	prob := 0.0
	m := (n / 2) + 1
	for k := m; k <= n; k++ {
		prob += dp[k]
	}

	return prob
}
