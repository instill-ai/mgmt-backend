package acl

// model
//   schema 1.1

// type visitor

// type user

// type code

// type organization
//   relations
//     define owner: [user]
//     define member: [user] or owner
//     define pending_owner: [user]
//     define pending_member: [user]
//     define can_create_organization: owner
//     define can_delete_organization: owner
//     define can_get_membership: owner or member
//     define can_remove_membership: owner
//     define can_set_membership: owner
//     define can_update_organization: owner

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

const ACLModel = `
  {
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
		  "owner": {
			"this": {}
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
					"relation": "owner"
				  }
				}
			  ]
			}
		  },
		  "pending_owner": {
			"this": {}
		  },
		  "pending_member": {
			"this": {}
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
					"relation": "member"
				  }
				}
			  ]
			}
		  },
		  "can_remove_membership": {
			"computedUserset": {
			  "object": "",
			  "relation": "owner"
			}
		  },
		  "can_set_membership": {
			"computedUserset": {
			  "object": "",
			  "relation": "owner"
			}
		  },
		  "can_update_organization": {
			"computedUserset": {
			  "object": "",
			  "relation": "owner"
			}
		  }
		},
		"metadata": {
		  "relations": {
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"member": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"pending_owner": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"pending_member": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"can_create_organization": {
			  "directly_related_user_types": []
			},
			"can_delete_organization": {
			  "directly_related_user_types": []
			},
			"can_get_membership": {
			  "directly_related_user_types": []
			},
			"can_remove_membership": {
			  "directly_related_user_types": []
			},
			"can_set_membership": {
			  "directly_related_user_types": []
			},
			"can_update_organization": {
			  "directly_related_user_types": []
			}
		  }
		}
	  },
	  {
		"type": "pipeline",
		"relations": {
		  "owner": {
			"this": {}
		  },
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					}
				  }
				}
			  ]
			}
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					}
				  }
				}
			  ]
			}
		  }
		},
		"metadata": {
		  "relations": {
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "organization"
				},
				{
				  "type": "user"
				}
			  ]
			},
			"admin": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"writer": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"executor": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "user",
				  "wildcard": {}
				},
				{
				  "type": "code"
				}
			  ]
			},
			"reader": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "user",
				  "wildcard": {}
				},
				{
				  "type": "code"
				},
				{
				  "type": "visitor",
				  "wildcard": {}
				}
			  ]
			}
		  }
		}
	  },
	  {
		"type": "connector",
		"relations": {
		  "owner": {
			"this": {}
		  },
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					}
				  }
				}
			  ]
			}
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
					"computedUserset": {
					  "object": "",
					  "relation": "member"
					},
					"tupleset": {
					  "object": "",
					  "relation": "owner"
					}
				  }
				}
			  ]
			}
		  }
		},
		"metadata": {
		  "relations": {
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "organization"
				},
				{
				  "type": "user"
				}
			  ]
			},
			"admin": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"writer": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"executor": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "user",
				  "wildcard": {}
				}
			  ]
			},
			"reader": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "user",
				  "wildcard": {}
				}
			  ]
			}
		  }
		}
	  }
	],
	"schema_version": "1.1"
  }
`
