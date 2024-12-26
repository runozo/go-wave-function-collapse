import sys
import json

# UP, RIGHT, DOWN, LEFT
TILEOPTIONS = {
	"tileGrass1.png":                     [0, 0, 0, 0], #  0 grass
	"tileGrass2.png":                     [0, 0, 0, 0],
	"tileGrass_roadCornerLL.png":         [0, 0, 1, 1], # 1 road with grass
	"tileGrass_roadCornerLR.png":         [0, 1, 1, 0],
	"tileGrass_roadCornerUL.png":         [1, 0, 0, 1],
	"tileGrass_roadCornerUR.png":         [1, 1, 0, 0],
	"tileGrass_roadCrossing.png":         [1, 1, 1, 1],
	"tileGrass_roadCrossingRound.png":    [1, 1, 1, 1],
	"tileGrass_roadEast.png":             [0, 1, 0, 1],
	"tileGrass_roadNorth.png":            [1, 0, 1, 0],
	"tileGrass_roadSplitE.png":           [1, 1, 1, 0],
	"tileGrass_roadSplitN.png":           [1, 1, 0, 1],
	"tileGrass_roadSplitS.png":           [0, 1, 1, 1],
	"tileGrass_roadSplitW.png":           [1, 0, 1, 1],
	"tileGrass_roadTransitionE.png":      [4, 3, 4, 1],
	"tileGrass_roadTransitionE_dirt.png": [4, 3, 4, 1],
	"tileGrass_roadTransitionN.png":      [3, 6, 1, 6],
	"tileGrass_roadTransitionN_dirt.png": [3, 6, 1, 6],
	"tileGrass_roadTransitionS.png":      [1, 8, 3, 8],
	"tileGrass_roadTransitionS_dirt.png": [1, 8, 3, 8],
	"tileGrass_roadTransitionW.png":      [5, 1, 5, 3],
	"tileGrass_roadTransitionW_dirt.png": [5, 1, 5, 3],
	"tileGrass_transitionE.png":          [4, 2, 4, 0],
	"tileGrass_transitionN.png":          [2, 6, 0, 6],
	"tileGrass_transitionS.png":          [0, 8, 2, 8],
	"tileGrass_transitionW.png":          [5, 0, 5, 2],
	"tileSand1.png":                      [2, 2, 2, 2],
	"tileSand2.png":                      [2, 2, 2, 2],
	"tileSand_roadCornerLL.png":          [2, 2, 3, 3],
	"tileSand_roadCornerLR.png":          [2, 3, 3, 2],
	"tileSand_roadCornerUL.png":          [3, 2, 2, 3],
	"tileSand_roadCornerUR.png":          [3, 3, 2, 2],
	"tileSand_roadCrossing.png":          [3, 3, 3, 3],
	"tileSand_roadCrossingRound.png":     [3, 3, 3, 3],
	"tileSand_roadEast.png":              [2, 3, 2, 3],
	"tileSand_roadNorth.png":             [3, 2, 3, 2],
	"tileSand_roadSplitE.png":            [3, 3, 3, 2],
	"tileSand_roadSplitN.png":            [3, 3, 2, 3],
	"tileSand_roadSplitS.png":            [2, 3, 3, 3],
	"tileSand_roadSplitW.png":            [3, 2, 3, 3],
}

# init options
mapped_options = {}
for key, options in TILEOPTIONS.items():
    mapped_options[key] = {
        "up": [],
        "right": [],
        "down": [],
        "left": []
    }
# map options
for key, options in TILEOPTIONS.items():
    up, right, down, left = options
    options_up = {}
    for key2, options2 in TILEOPTIONS.items():
        key2 = key2.split(".")[0]
        up2, right2, down2, left2 = options2
        if up == down2:
            mapped_options[key]["up"].append(key2)
        if right == left2:
            mapped_options[key]["right"].append(key2)
        if down == up2:
            mapped_options[key]["down"].append(key2)
        if left == right2:
            mapped_options[key]["left"].append(key2)

# print(mapped_options)

tiles = []
for key, options in mapped_options.items():
    tiles.append({
        "name": key.split(".")[0],
        "image_name": key,
        "options": options,
        "type": "ground"
    })

import random
import xml.etree.ElementTree as ET
tree = ET.parse('./allSprites_default.xml')
root = tree.getroot()
for child in root:
    print(child.tag, child.attrib['name'])
    key = child.attrib['name'].split(".")[0]
    if [_ for _ in tiles if _["name"] == key]:
        # it's a ground tile
        for tile in tiles:
            if tile["name"] == key:
                tile.update({
                    "x": int(child.attrib["x"]),
                    "y": int(child.attrib["y"]),
                    "width": int(child.attrib["width"]),
                    "height": int(child.attrib["height"]),
                    "weight": int(child.attrib.get("weight", "20")),
                })
                tile["options"].update({
                    "above": [],
                    "below": []
                })
    else:
        # it's a random tile
        tiles.append({
            "name": key,
            "image_name": child.attrib['name'],
            "options": {
                "up": [],
                "right": [],
                "down": [],
                "left": [],
                "above": [],
                "below": [],
            },
            "x": int(child.attrib["x"]),
            "y": int(child.attrib["y"]),
            "width": int(child.attrib["width"]),
            "height": int(child.attrib["height"]),
            "weight": int(child.attrib.get("weight", random.randint(0, 100))),
            "type": "random"
        })

#import json
#with open("mapped_tiles.json", "w") as f:
#    json.dump(tiles, f, indent=4)

#import xmltodict
#with open("mapped_tiles.xml", "w") as f:
#    xml = xmltodict.unparse({"tiles": tiles}, pretty=True)
#    f.write(xml)

# print(open("mapped_tiles.xml").read())
# print(open("mapped_tiles.json").read())
json.dump(tiles, sys.stdout, indent=4)