package acl

const ACLModel = `
{
	"type_definitions": [
	  {
		"type": "user",
		"relations": {},
		"metadata": null
	  },
	  {
		"type": "team",
		"relations": {
		  "member": {
			"this": {}
		  }
		},
		"metadata": {
		  "relations": {
			"member": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "team",
				  "relation": "member"
				}
			  ]
			}
		  }
		}
	  },
	  {
		"type": "organization",
		"relations": {
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
		  "owner": {
			"this": {}
		  },
		  "repo_admin": {
			"this": {}
		  },
		  "repo_reader": {
			"this": {}
		  },
		  "repo_writer": {
			"this": {}
		  }
		},
		"metadata": {
		  "relations": {
			"member": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"owner": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				}
			  ]
			},
			"repo_admin": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "organization",
				  "relation": "member"
				}
			  ]
			},
			"repo_reader": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "organization",
				  "relation": "member"
				}
			  ]
			},
			"repo_writer": {
			  "directly_related_user_types": [
				{
				  "type": "user"
				},
				{
				  "type": "organization",
				  "relation": "member"
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
