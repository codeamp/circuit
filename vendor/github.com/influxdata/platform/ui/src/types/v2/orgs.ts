export interface Organization {
  id: string
  name: string
  links: OrgLinks
}

interface OrgLinks {
  buckets: string
  dashboards: string
  members: string
  self: string
  tasks: string
}
