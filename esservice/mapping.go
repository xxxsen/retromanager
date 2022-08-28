package esservice

var IndexGameInfoTabMapping = `{
	"mappings": {
	  "properties": {
		"content": {
		  "type": "text",
		  "analyzer": "ik_max_word",
		  "search_analyzer": "ik_smart"
		},
		"create_time": {
		  "type": "long"
		},
		"desc": {
		  "type": "text",
		  "fields": {
			"keyword": {
			  "type": "keyword",
			  "ignore_above": 256
			}
		  }
		},
		"display_name": {
		  "type": "text",
		  "fields": {
			"keyword": {
			  "type": "keyword",
			  "ignore_above": 256
			}
		  }
		},
		"down_key": {
		  "type": "text",
		  "fields": {
			"keyword": {
			  "type": "keyword",
			  "ignore_above": 256
			}
		  }
		},
		"extinfo": {
		  "properties": {
			"developer": {
			  "type": "text",
			  "fields": {
				"keyword": {
				  "type": "keyword",
				  "ignore_above": 256
				}
			  }
			},
			"image": {
			  "type": "text",
			  "fields": {
				"keyword": {
				  "type": "keyword",
				  "ignore_above": 256
				}
			  }
			},
			"players": {
			  "type": "long"
			},
			"publisher": {
			  "type": "text",
			  "fields": {
				"keyword": {
				  "type": "keyword",
				  "ignore_above": 256
				}
			  }
			},
			"rating": {
			  "type": "long"
			},
			"releasedate": {
			  "type": "text",
			  "fields": {
				"keyword": {
				  "type": "keyword",
				  "ignore_above": 256
				}
			  }
			},
			"video": {
			  "type": "text",
			  "fields": {
				"keyword": {
				  "type": "keyword",
				  "ignore_above": 256
				}
			  }
			}
		  }
		},
		"file_name": {
		  "type": "text",
		  "fields": {
			"keyword": {
			  "type": "keyword",
			  "ignore_above": 256
			}
		  }
		},
		"file_size": {
		  "type": "long"
		},
		"hash": {
		  "type": "text",
		  "fields": {
			"keyword": {
			  "type": "keyword",
			  "ignore_above": 256
			}
		  }
		},
		"id": {
		  "type": "long"
		},
		"platform": {
		  "type": "long"
		},
		"update_time": {
		  "type": "long"
		}
	  }
	}
  }`
