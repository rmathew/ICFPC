import sys
import utils

UNKNOWN_PUNTER_ID = -1
INVALID_SITE_ID = -1

class Site():
    def __init__(self):
        self.x = -1
        self.y = -1
        self.neighbors_list = []

class River():
    def __init__(self, source, target):
        self.source = min(source, target)
        self.target = max(source, target)

    def __str__(self):
        return "(%d,%d)" % (self.source, self.target)

    def __hash__(self):
        return hash(self.source * 1024 + self.target)

    def __eq__(self, rhs):
        return (isinstance(rhs, River) and self.source == rhs.source and
            self.target == rhs.target)

INVALID_RIVER = River(INVALID_SITE_ID, INVALID_SITE_ID)

class WorldMap():

    def __init__(self, world_map_dict):
        self.sites_dict = {}
        self.mines_set = {}
        if "sites" in world_map_dict:
            for a_site_dict in world_map_dict["sites"]:
                the_site = Site()
                the_site.x = a_site_dict.get("x", -1)
                the_site.y = a_site_dict.get("y", -1)
                self.sites_dict[int(a_site_dict["id"])] = the_site
        if "rivers" in world_map_dict:
            for a_river_dict in world_map_dict["rivers"]:
                src_site_id = int(a_river_dict["source"])
                tgt_site_id = int(a_river_dict["target"])
                src_site = self.sites_dict[src_site_id]
                src_site.neighbors_list.append(tgt_site_id)
                tgt_site = self.sites_dict[tgt_site_id]
                tgt_site.neighbors_list.append(src_site_id)
        if "mines" in world_map_dict:
            mines_list = []
            for a_mine_id in world_map_dict["mines"]:
                mines_list.append(int(a_mine_id))
            self.mines_set = frozenset(mines_list)

    def get_num_sites(self):
        return len(self.sites_dict)

    def get_num_rivers(self):
        num_rivers = 0
        for a_site_id, a_site in self.sites_dict.items():
            for a_neighbor_id in a_site.neighbors_list:
                num_rivers += 1
        return num_rivers / 2

    def get_num_mines(self):
        return len(self.mines_set)

    def to_dict(self):
        return_dict = {}
        sites_list = []
        rivers_list = []
        for a_site_id, a_site in self.sites_dict.items():
            sites_list.append({"id": a_site_id})
            for a_neighbor_id in a_site.neighbors_list:
                if a_site_id > a_neighbor_id:
                    continue
                rivers_list.append(
                    {"source": a_site_id, "target": a_neighbor_id})
        return_dict["sites"] = sites_list
        return_dict["rivers"] = rivers_list
        mines_list = []
        for a_mine_id in self.mines_set:
            mines_list.append(a_mine_id)
        return_dict["mines"] = mines_list
        return return_dict

class WorldState():

    def __init__(self, world_dict):
        self.world_map = WorldMap(world_dict["map"])
        self.num_punters = int(world_dict["punters"])
        self.claims_dict = {}
        for a_claim_dict in world_dict["claims"]:
            punter_id = int(a_claim_dict["punter"])
            river = River(
                int(a_claim_dict["source"]), int(a_claim_dict["target"]))
            is_valid_claim = self.add_punter_claim(punter_id, river)
            if not is_valid_claim:
              utils.eprint("ERROR: world_dict has an invalid claim %d=%s" %
                  (punter_id, river))

    def to_dict(self):
        return_dict = {}
        return_dict["map"] = self.world_map.to_dict()
        return_dict["punters"] = int(self.num_punters)
        claims_list = []
        for a_river, a_punter_id in self.claims_dict.items():
            claims_list.append({"punter": a_punter_id,
                "source": a_river.source, "target": a_river.target})
        return_dict["claims"] = claims_list
        return return_dict

    def add_punter(self):
        self.num_punters += 1
        return self.num_punters - 1

    def get_claiming_punter(self, river):
        return self.claims_dict.get(river, UNKNOWN_PUNTER_ID)

    def add_punter_claim(self, punter_id, river):
        if river == INVALID_RIVER:
            return False
        if river in self.claims_dict:
            return False
        self.claims_dict[river] = punter_id
        return True

    def calculate_score(self, punter_id):
        accessible_edges_dict = self._get_accessible_edges(punter_id)
        score = 0
        for a_mine_id in self.world_map.mines_set:
            score += self._get_score_for_mine(a_mine_id, accessible_edges_dict)
        return score

    def _get_accessible_edges(self, punter_id):
        nodes_to_edges_dict = {}
        for a_site_id, a_site in self.world_map.sites_dict.items():
            site_edges_list = []
            for a_neighbor_id in a_site.neighbors_list:
                a_river = River(a_site_id, a_neighbor_id)
                if self.get_claiming_punter(a_river) == punter_id:
                    site_edges_list.append(a_neighbor_id)
            nodes_to_edges_dict[a_site_id] = site_edges_list
        return nodes_to_edges_dict

    def _get_score_for_mine(self, a_mine_id, accessible_edges_dict):
        if len(accessible_edges_dict[a_mine_id]) == 0:
            return 0

        # Use Dijkstra's algorithm to get the shortest paths from the mine to
        # all the other sites _in the original graph_.
        shortest_dist_dict = self._get_shortest_distances(a_mine_id)

        # Find all the sites that can be visited from the mine _in the
        # accessible graph_.
        accessible_sites_set = set()
        self._fill_accessible_sites(a_mine_id, accessible_edges_dict,
            accessible_sites_set)

        score = 0
        for a_site in accessible_sites_set:
            shortest_dist = shortest_dist_dict[a_site]
            score += shortest_dist * shortest_dist
        return score

    def _get_shortest_distances(self, a_mine_id):
        orig_sites_dict = self.world_map.sites_dict
        MAX_POSSIBLE_DIST = sys.maxsize
        seen_set = set()
        dist_dict = {a_mine_id: 0}
        curr_site_id = a_mine_id
        dist_to_curr_site = 0
        while not curr_site_id in seen_set:
            seen_set.add(curr_site_id)
            for an_edge in orig_sites_dict[curr_site_id].neighbors_list:
                new_dist_to_edge = dist_to_curr_site + 1
                curr_dist_to_edge = dist_dict.get(an_edge, MAX_POSSIBLE_DIST)
                if curr_dist_to_edge > new_dist_to_edge:
                    dist_dict[an_edge] = new_dist_to_edge

            dist_to_curr_site = MAX_POSSIBLE_DIST
            for a_site_id in orig_sites_dict.keys():
                if a_site_id in seen_set:
                    continue
                dist_to_site = dist_dict.get(a_site_id, MAX_POSSIBLE_DIST)
                if dist_to_curr_site > dist_to_site:
                    dist_to_curr_site = dist_to_site
                    curr_site_id = a_site_id
        return dist_dict

    def _fill_accessible_sites(self, src, accessible_edges, accessible_sites):
        accessible_sites.add(src)
        for tgt in accessible_edges[src]:
            if tgt in accessible_sites:
                continue
            self._fill_accessible_sites(tgt, accessible_edges, accessible_sites)
