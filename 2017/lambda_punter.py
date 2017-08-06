import sys
import utils
import lambda_world

INVALID_SITE_ID = -1

class LambdaPunter():

    def __init__(self, name):
        self.punter_name = name
        self.punter_id = lambda_world.UNKNOWN_PUNTER_ID
        self.world_state = None

    def get_river_to_claim(self):
        return (INVALID_SITE_ID, INVALID_SITE_ID)

    def shake_hands(self):
        self.send_msg({"me": self.punter_name})
        response_dict = self.recv_msg()
        if response_dict.get("you", "<<UNKNOWN>>") != self.punter_name:
            utils.eprint("WARNING: Botched hand-shake.")

    def set_up(self, cmd_dict):
        utils.eprint("INFO: Hello, my name is '%s'." % self.punter_name)
        self.punter_id = int(cmd_dict["punter"])
        num_punters = int(cmd_dict["punters"])
        utils.eprint("INFO: I am the Punter with id %d among %d Punters." %
            (self.punter_id, num_punters))
        world_state_dict = {}
        world_state_dict["map"] = cmd_dict["map"]
        world_state_dict["punters"] = num_punters
        world_state_dict["claims"] = []
        self.world_state = lambda_world.WorldState(world_state_dict)
        self.send_msg({"ready": self.punter_id, "state": self.get_state_dict()})

    def make_a_move(self, cmd_dict):
        self.update_state(cmd_dict["state"], cmd_dict["move"]["moves"])
        claimed_river = self.get_river_to_claim()
        utils.eprint("INFO: My move is %s." % self.claim_to_str(self.punter_id,
            claimed_river[0], claimed_river[1]))
        response_dict = {"state": self.get_state_dict()}
        if claimed_river[0] == INVALID_SITE_ID:
            response_dict["pass"] = {"punter": self.punter_id}
        else:
            response_dict["claim"] = {"punter": self.punter_id,
                "source": claimed_river[0], "target": claimed_river[1]}
        self.send_msg(response_dict)

    def wrap_it_up(self, cmd_dict):
        stop_dict = cmd_dict["stop"]
        scores_list = stop_dict["scores"]
        my_score = -1
        for a_punters_score_dict in scores_list:
            if a_punters_score_dict["punter"] == self.punter_id:
                my_score = a_punters_score_dict["score"]
                break
        utils.eprint("INFO: The game has ended and my score is %d." % my_score)

    def punt(self):
        self.shake_hands()
        cmd_dict = self.recv_msg()
        if "punter" in cmd_dict:
            self.set_up(cmd_dict)
        elif "move" in cmd_dict:
            self.make_a_move(cmd_dict)
        elif "stop" in cmd_dict:
            self.wrap_it_up(cmd_dict)
        else:
            utils.eprint("ERROR: Unknown server-command.")
            utils.eprint(cmd_dict)

    def send_msg(self, obj):
        utils.encode_obj(sys.stdout, obj)

    def recv_msg(self):
        return utils.decode_obj(sys.stdin)

    def claim_to_str(self, punter_id, source, target):
        if source == INVALID_SITE_ID:
            return "%d=P" % punter_id
        return "%d=C(%d,%d)" % (punter_id, source, target)

    def update_state(self, state_dict, prev_moves_list):
        self.punter_id = int(state_dict["punter"])
        self.world_state = lambda_world.WorldState(state_dict["world_state"])
        prev_moves_str = ""
        for a_move_dict in prev_moves_list:
            if "pass" in a_move_dict:
                prev_moves_str += self.claim_to_str(
                    int(a_move_dict["pass"]["punter"]), INVALID_SITE_ID,
                    INVALID_SITE_ID)
            else:
                claim_dict = a_move_dict["claim"]
                punter_id = int(claim_dict["punter"])
                source = int(claim_dict["source"])
                target = int(claim_dict["target"])
                move_str = self.claim_to_str(punter_id, source, target)
                claim_status = self.world_state.add_punter_claim(
                    punter_id, source, target)
                if not claim_status:
                    utils.eprint("WARNING: Got invalid move %s." % move_str)
                prev_moves_str += move_str
            prev_moves_str += " "
        utils.eprint("INFO: Previous moves:\n  %s" % prev_moves_str)

    def get_state_dict(self):
        return {"punter": int(self.punter_id),
            "world_state": self.world_state.to_dict()}

class NaivePunter(LambdaPunter):

    def __init__(self, name):
        LambdaPunter.__init__(self, name)
