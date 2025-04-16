# def all_unique(lst):
#     return len(lst) == len(set(lst))

# def is_eligible(lst):
#     p = 1.
#     for x in lst:
#         p*=x
#     print(lst, "->", p)
#     return p>600.

# firstEligible = (10, 9, 8, 7)
# n = 15


# eligibles_neigh = [firstEligible]
# size = len(firstEligible)

# # permutate
# i = 0
# while i<len(eligibles_neigh):
#     el = eligibles_neigh[i]
#     print(i, ":", el)
#     for idx in range(size):
#         for jdx in range(size):
#             if idx==jdx:
#                 continue
#             newT = list(el)
#             newT[idx] += 1
#             newT[jdx] -= 1
#             newT.sort(reverse=True)
#             newT = tuple(newT)
#             print("newtT=", newT)
#             if all_unique(newT) and newT not in eligibles_neigh:
#                 if  is_eligible(newT):
#                     print("appending", newT)
#                     eligibles_neigh.append(tuple(newT))
#     i+=1

# # all greater
# eligibles = set(eligibles_neigh)

# # Print
# i = 10
# for e in eligibles:
#     i-=1
#     if i>0:
#         print(e, end="\t")
#     else:
#         print(e)
#         i = 10

# print("\n\n\n")

# for el in eligibles_neigh:
#     newT = list(el[:])
#     while newT[0] <= n:
#         newT[-1] += 1
#         if newT[-1] > n:
#             idx = size-1
#             while idx>0 and newT[idx] > n:
#                 newT[idx] = el[idx]
#                 newT[idx-1] += 1
#                 idx -= 1
        
#         if all_unique(newT):
#             eligibles.add(tuple(newT))

        


# # Print
# i = 10
# for e in eligibles:
#     i-=1
#     if i>0:
#         print(e, end="\t")
#     else:
#         print(e)
#         i = 10

import os

d = os.getcwd()
os.
print ("d is  ", d)
