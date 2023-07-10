# Doubletree Traceroute in Go

Doubletree for distributed network monitors
Probe uses a modified version of the original Go traceroute by Aeden on Github.
Original traceroute [github.com/aeden/traceroute]
Package docs for Aeden traceroute [https://pkg.go.dev/github.com/aeden/traceroute]

## Works Cited

Doubletree

* Donnet, B., Raoult, P., Friedman, T., & Crovella, M. (2006). Deployment of an Algorithm for Large-Scale Topology Discovery. IEEE JOURNAL ON SELECTED AREAS IN COMMUNICATIONS, (6th IEEE International Conference on IP Operations and Management). https://doi.org/10.1007/11908852_17 
* 

## Notes
* On the reliability of the bloom filter data structure
    * According to the paper, let us limit the maximum size transmission of the stop set to around 15 kb.
    * Since bloom filters only hold a single bit per index, we must accept a certain probability of collision. In regards to our project, a collision would mean that a node stops probing prematurely, believing an unseen interface to be already seen. In regards to the project, this is not a critical failure.
    * For the sake of this project, I accept a collision probability of 0.005.
    * There exists an upper bound to the number of seen stops belonging to a range of IPs, based on the span of the network. The second Doubletree paper found shared nodes frequently occur around 5 to 7 hops along a path.
        * Therefore, given a slice of 500 IP addresses assigned to a probe worker, the stop set, once filled, would likely contain around 2,500 to 3,500 (hop, destination) pairs. We round this up to 4,000 for safety.
    * With 4,000 items to be stored in a bloom filter, 2 hash functions, and a maximum accepted collision probability of .005, the resulting bloom filter requires 13.32 KB.
    * This fits within the requirements of the Doubletree paper, and we implement this using our bit set data structure.
* [a blog post](https://medium.com/@val_deleplace/7-ways-to-implement-a-bit-set-in-go-91650229b386) about different strategies for implementing bitsets in go. one structure consists of a struct containing an array of uint64s (since Go does not allow for raw binary operations on arbitrary data types)
    * Taking the union of two such bitsets would simply require an OR operation on each of the corresponding slices
    * Since our bloom filter requires ~14 kb, this would give us an array of length 220. This is quite a manageable array.
* The size of a bitset encompassing all of IPV4 space is 2^32 bits or roughly half a gigabyte, where each IP address hashes directly to itself. This is large but easy for a server to maintain, if it needs to register all IPs it has ever seen, etc.

* set.go based on this tutorial: https://www.davidkaya.com/sets-in-golang/

* I revisit this from time to time
