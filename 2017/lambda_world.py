UNKNOWN_PUNTER_ID = -1

class Site():
    def __init__(self):
        self.x = -1
        self.y = -1
        self.neighbors_list = []

class WorldMap():

    def __init__(self, world_map_dict):
        self.sites_dict = {}
        self.mines_list = []
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
            for a_mine_id in world_map_dict["mines"]:
                self.mines_list.append(int(a_mine_id))

    def get_num_sites(self):
        return len(self.sites_dict)

    def get_num_rivers(self):
        num_rivers = 0
        for a_site_id, a_site in self.sites_dict.items():
            for a_neighbor_id in a_site.neighbors_list:
                if a_site_id < a_neighbor_id:
                    num_rivers += 1
        return num_rivers

    def get_num_mines(self):
        return len(self.mines_list)

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
        for a_mine_id in self.mines_list:
            mines_list.append(a_mine_id)
        return_dict["mines"] = mines_list
        return return_dict

class WorldState():

    def __init__(self, world_dict):
        self.world_map = WorldMap(world_dict["map"])
        self.num_punters = int(world_dict["punters"])
        self.claims_dict = {}
        claims_list = world_dict["claims"]
        for a_claim_dict in claims_list:
            self.add_punter_claim(int(a_claim_dict["punter"]),
                int(a_claim_dict["source"]), int(a_claim_dict["target"]))

    def to_dict(self):
        return_dict = {}
        return_dict["map"] = self.world_map.to_dict()
        return_dict["punters"] = int(self.num_punters)
        claims_list = []
        for a_river, a_punter_id in self.claims_dict.items():
            claims_list.append({"punter": a_punter_id, "source": a_river[0],
                "target": a_river[1]})
        return_dict["claims"] = claims_list
        return return_dict

    def add_punter(self):
        self.num_punters += 1
        return self.num_punters - 1

    def get_claiming_punter(self, source, target):
        river = (min(source, target), max(source, target))
        self.claims_dict.get(river, UNKNOWN_PUNTER_ID)

    def add_punter_claim(self, punter_id, source, target):
        river = (min(source, target), max(source, target))
        if river in self.claims_dict:
            return False
        self.claims_dict[river] = punter_id
        return True
