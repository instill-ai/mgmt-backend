package acl

// model
//   schema 1.1

// type visitor

// type user

// type code

// type organization
//   relations
//     define owner: [user]
//     define admin: [user] or owner
//     define member: [user] or admin or owner
//     define pending_owner: [user]
//     define pending_admin: [user]
//     define pending_member: [user]
//     define can_create_organization: owner
//     define can_delete_organization: owner
//     define can_get_membership: owner or admin or member
//     define can_remove_membership: owner or admin
//     define can_set_membership: owner or admin
//     define can_update_organization: owner or admin

// type pipeline
//   relations
//   define owner: [organization, user]
//     define admin: [user] or owner or member from owner
//     define writer: [user] or admin or member from owner
//     define executor: [user, user:*, code] or writer or member from owner
//     define reader: [user, user:*, code, visitor:*] or executor or member from owner

// type connector
//   relations
//     define owner: [organization, user]
//     define admin: [user] or owner or member from owner
//     define writer: [user] or admin or member from owner
//     define executor: [user, user:*] or writer or member from owner
//     define reader: [user, user:*] or executor or member from owner

// type model_
//   relations
//   define owner: [organization, user]
//     define admin: [user] or owner or member from owner
//     define writer: [user] or admin or member from owner
//     define executor: [user, user:*, code] or writer or member from owner
//     define reader: [user, user:*, code, visitor:*] or executor or member from owner

const ACLModel = `
  {
	"schema_version": "1.1",
	"type_definitions": [
	  {
		"type": "visitor",
		"relations": {},
		"metadata": null
	  },
	  {
		"type": "user",
		"relations": {},
		"metadata": null
	  },
	  {
		"type": "code",
		"relations": {},
		"metadata": null
	  },
	  {
		"type": "organization",
		"relations": {
		  "admin": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				}
			  ]
			}
		  },
		  "can_create_organization": {
			"computedUserset": {
			  "object": "",
			  "relation": "owner"
			}
		  },
		  "can_delete_organization": {
			"computedUserset": {
			  "object": "",
			  "relation": "owner"
			}
		  },
		  "can_get_membership": {
			"union": {
			  "child": [
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "member"
				  }
				}
			  ]
			}
		  },
		  "can_remove_membership": {
			"union": {
			  "child": [
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				}
			  ]
			}
		  },
		  "can_set_membership": {
			"union": {
			  "child": [
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				}
			  ]
			}
		  },
		  "can_update_organization": {
			"union": {
			  "child": [
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				}
			  ]
			}
		  },
		  "member": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				}
			  ]
			}
		  },
		  "owner": {
			"this": {}
		  },
		  "pending_admin": {
			"this": {}
		  },
		  "pending_member": {
			"this": {}
		  },
		  "pending_owner": {
			"this": {}
		  }
		},
		"metadata": {
		  "relations": {
			"admin": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"can_create_organization": {
			  "directly_related_user_types": [],
			  "module": "",
			  "source_info": null
			},
			"can_delete_organization": {
			  "directly_related_user_types": [],
			  "module": "",
			  "source_info": null
			},
			"can_get_membership": {
			  "directly_related_user_types": [],
			  "module": "",
			  "source_info": null
			},
			"can_remove_membership": {
			  "directly_related_user_types": [],
			  "module": "",
			  "source_info": null
			},
			"can_set_membership": {
			  "directly_related_user_types": [],
			  "module": "",
			  "source_info": null
			},
			"can_update_organization": {
			  "directly_related_user_types": [],
			  "module": "",
			  "source_info": null
			},
			"member": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"pending_admin": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"pending_member": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"pending_owner": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			}
		  },
		  "module": "",
		  "source_info": null
		}
	  },
	  {
		"type": "pipeline",
		"relations": {
		  "admin": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "executor": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "writer"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "owner": {
			"this": {}
		  },
		  "reader": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "executor"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "writer": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  }
		},
		"metadata": {
		  "relations": {
			"admin": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"executor": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				},
				{
				  "type": "user",
				  "wildcard": {},
				  "condition": ""
				},
				{
				  "type": "code",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "organization",
				  "condition": ""
				},
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"reader": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				},
				{
				  "type": "user",
				  "wildcard": {},
				  "condition": ""
				},
				{
				  "type": "code",
				  "condition": ""
				},
				{
				  "type": "visitor",
				  "wildcard": {},
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"writer": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			}
		  },
		  "module": "",
		  "source_info": null
		}
	  },
	  {
		"type": "connector",
		"relations": {
		  "admin": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "executor": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "writer"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "owner": {
			"this": {}
		  },
		  "reader": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "executor"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "writer": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  }
		},
		"metadata": {
		  "relations": {
			"admin": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"executor": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				},
				{
				  "type": "user",
				  "wildcard": {},
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "organization",
				  "condition": ""
				},
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"reader": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				},
				{
				  "type": "user",
				  "wildcard": {},
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"writer": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			}
		  },
		  "module": "",
		  "source_info": null
		}
	  },
	  {
		"type": "model_",
		"relations": {
		  "admin": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "owner"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "executor": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "writer"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "owner": {
			"this": {}
		  },
		  "reader": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "executor"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  },
		  "writer": {
			"union": {
			  "child": [
				{
				  "this": {}
				},
				{
				  "computedUserset": {
					"object": "",
					"relation": "admin"
				  }
				},
				{
				  "tupleToUserset": {
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					},
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					}
				  }
				}
			  ]
			}
		  }
		},
		"metadata": {
		  "relations": {
			"admin": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"executor": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				},
				{
				  "type": "user",
				  "wildcard": {},
				  "condition": ""
				},
				{
				  "type": "code",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "organization",
				  "condition": ""
				},
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"reader": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				},
				{
				  "type": "user",
				  "wildcard": {},
				  "condition": ""
				},
				{
				  "type": "code",
				  "condition": ""
				},
				{
				  "type": "visitor",
				  "wildcard": {},
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			},
			"writer": {
			  "directly_related_user_types": [
				{
				  "type": "user",
				  "condition": ""
				}
			  ],
			  "module": "",
			  "source_info": null
			}
		  },
		  "module": "",
		  "source_info": null
		}
	  }
	],
	"conditions": {}
  }
`
