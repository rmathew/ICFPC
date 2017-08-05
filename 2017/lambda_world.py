from collections import namedtuple

Site = namedtuple("Site", "x y")
River = namedtuple("River", "src tgt")

INVALID_RIVER = River(src=-1, tgt=-1)

class WorldMap():

    def __init__(self, world_map_dict):
        self.sites = {}
        self.rivers = []
        self.mines = []
        if "sites" in world_map_dict:
            for a_site_dict in world_map_dict["sites"]:
                self.sites[a_site_dict["id"]] = Site(
                    x=a_site_dict.get("x", 0), y=a_site_dict.get("y", 0))
        if "rivers" in world_map_dict:
            for a_river_dict in world_map_dict["rivers"]:
                self.rivers.append(River(
                    src=a_river_dict["source"], tgt=a_river_dict["target"]))
        if "mines" in world_map_dict:
            for a_mine_id in world_map_dict["mines"]:
                self.mines.append(a_mine_id)

    def to_dict(self):
        dict = {}
        sites_list = []
        for a_site_id, a_site in self.sites.items():
            sites_list.append({"id": a_site_id, "x": a_site.x, "y": a_site.y})
        dict["sites"] = sites_list
        rivers_list = []
        for a_river in self.rivers:
            rivers_list.append({"source": a_river.src, "target": a_river.tgt})
        dict["rivers"] = rivers_list
        mines_list = []
        for a_mine_id in self.mines:
            mines_list.append(a_mine_id)
        dict["mines"] = mines_list
        return dict

class Punter():

    def __init__(self, name):
        self.name = name
        self.river_last_claimed = INVALID_RIVER

class World():

    def __init__(self, world_map):
        self.world_map = world_map
        self.punters = []

    def add_punter(self, name):
        self.punters.append(Punter(name))
