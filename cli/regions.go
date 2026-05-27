package main

type Region struct {
	Name   string
	Code   string
	Group  string
	Beacon string
}

type RegionGroup struct {
	Label   string
	Servers []Region
}

var regionGroups = []RegionGroup{
	{
		Label: "EUROPE",
		Servers: []Region{
			{Name: "Ireland", Code: "eu-west-1", Group: "EUROPE", Beacon: "gamelift-ping.eu-west-1.api.aws:7770"},
			{Name: "London", Code: "eu-west-2", Group: "EUROPE", Beacon: "gamelift-ping.eu-west-2.api.aws:7770"},
			{Name: "Frankfurt", Code: "eu-central-1", Group: "EUROPE", Beacon: "gamelift-ping.eu-central-1.api.aws:7770"},
		},
	},
	{
		Label: "NORTH AMERICA",
		Servers: []Region{
			{Name: "Ohio", Code: "us-east-2", Group: "NORTH AMERICA", Beacon: "gamelift-ping.us-east-2.api.aws:7770"},
			{Name: "Oregon", Code: "us-west-2", Group: "NORTH AMERICA", Beacon: "gamelift-ping.us-west-2.api.aws:7770"},
			{Name: "N. California", Code: "us-west-1", Group: "NORTH AMERICA", Beacon: "gamelift-ping.us-west-1.api.aws:7770"},
			{Name: "Montreal", Code: "ca-central-1", Group: "NORTH AMERICA", Beacon: "gamelift-ping.ca-central-1.api.aws:7770"},
		},
	},
	{
		Label: "SOUTH AMERICA",
		Servers: []Region{
			{Name: "Sao Paulo", Code: "sa-east-1", Group: "SOUTH AMERICA", Beacon: "gamelift-ping.sa-east-1.api.aws:7770"},
		},
	},
	{
		Label: "ASIA PACIFIC",
		Servers: []Region{
			{Name: "Tokyo", Code: "ap-northeast-1", Group: "ASIA PACIFIC", Beacon: "gamelift-ping.ap-northeast-1.api.aws:7770"},
			{Name: "Seoul", Code: "ap-northeast-2", Group: "ASIA PACIFIC", Beacon: "gamelift-ping.ap-northeast-2.api.aws:7770"},
			{Name: "Singapore", Code: "ap-southeast-1", Group: "ASIA PACIFIC", Beacon: "gamelift-ping.ap-southeast-1.api.aws:7770"},
			{Name: "Mumbai", Code: "ap-south-1", Group: "ASIA PACIFIC", Beacon: "gamelift-ping.ap-south-1.api.aws:7770"},
		},
	},
	{
		Label: "OCEANIA",
		Servers: []Region{
			{Name: "Sydney", Code: "ap-southeast-2", Group: "OCEANIA", Beacon: "gamelift-ping.ap-southeast-2.api.aws:7770"},
		},
	},
}

func totalRegions() int {
	n := 0
	for _, g := range regionGroups {
		n += len(g.Servers)
	}
	return n
}
