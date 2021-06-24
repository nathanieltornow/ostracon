package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

type Shard struct {
	T        string        `yaml:"type"`
	Color    int64         `yaml:"color"`
	IP       string        `yaml:"ip"`
	ParentIP string        `yaml:"parent_ip"`
	Interval time.Duration `yaml:"interval"`
	Disk     string        `yaml:"disk"`
	Root     bool          `yaml:"root"`
}

type Config struct {
	Shards []*Shard `yaml:"shards"`
}

func ParseConfig() (*Config, error) {
	buf, err := ioutil.ReadFile("cluster.config.yaml")
	if err != nil {
		return nil, err
	}
	c := Config{}
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (shard *Shard) getColors(ipToShardMap map[string]*Shard) []int64 {
	res := make([]int64, 0)
	parent := ipToShardMap[shard.ParentIP]
	for parent != nil {
		res = append(res, parent.Color)
		parent = ipToShardMap[parent.ParentIP]
	}
	return res
}

func GetWriteShards(color int64) ([]string, error) {
	readShards, err := GetReadShards(color)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, l := range readShards {
		res = append(res, l...)
	}
	return res, nil
}

func GetReadShards(color int64) ([][]string, error) {
	conf, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	res := make([][]string, 0)
	groupLeaders := make(map[string]int)
	ipToShardMap := ipToShard(conf.Shards)
	for _, shard := range conf.Shards {
		if shard.T != "record" {
			continue
		}
		colors := shard.getColors(ipToShardMap)
		for _, c := range colors {
			if color != c {
				continue
			}
			ind, ok := groupLeaders[shard.ParentIP]
			if !ok {
				groupLeaders[shard.ParentIP] = len(res)
				res = append(res, []string{shard.IP})
				continue
			}
			res[ind] = append(res[ind], shard.IP)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("error: color isn't supported by this cluster")
	}

	return res, nil
}

func ipToShard(shards []*Shard) map[string]*Shard {
	resMap := make(map[string]*Shard)
	for _, s := range shards {
		resMap[s.IP] = s
	}
	return resMap
}
