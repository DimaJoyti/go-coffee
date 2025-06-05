'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Database, 
  Search, 
  Key, 
  Hash, 
  List, 
  Users, 
  BarChart3,
  Clock,
  Trash2,
  Eye,
  RefreshCw,
  Filter,
  Download
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useRedisData } from '@/hooks/use-redis'
import { cn } from '@/lib/utils'

interface RedisKey {
  key: string
  type: string
  ttl: number
  memory_usage?: number
  length?: number
  cardinality?: number
  field_count?: number
}

interface RedisExplorerProps {
  className?: string
}

export function RedisExplorer({ className }: RedisExplorerProps) {
  const [activeTab, setActiveTab] = useState('keys')
  const [searchPattern, setSearchPattern] = useState('*')
  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [dataType, setDataType] = useState('all')
  const [limit, setLimit] = useState(100)
  
  const { 
    keys, 
    keyDetails, 
    isLoading, 
    error, 
    exploreKeys, 
    getKeyDetails,
    exploreData 
  } = useRedisData()

  useEffect(() => {
    exploreKeys({ pattern: searchPattern, limit })
  }, [searchPattern, limit, exploreKeys])

  const handleKeySelect = async (key: string) => {
    setSelectedKey(key)
    await getKeyDetails(key)
  }

  const handleSearch = () => {
    exploreKeys({ pattern: searchPattern, limit })
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'string':
        return <Key className="h-4 w-4" />
      case 'hash':
        return <Hash className="h-4 w-4" />
      case 'list':
        return <List className="h-4 w-4" />
      case 'set':
        return <Users className="h-4 w-4" />
      case 'zset':
        return <BarChart3 className="h-4 w-4" />
      default:
        return <Database className="h-4 w-4" />
    }
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'string':
        return 'bg-blue-100 text-blue-800'
      case 'hash':
        return 'bg-green-100 text-green-800'
      case 'list':
        return 'bg-purple-100 text-purple-800'
      case 'set':
        return 'bg-orange-100 text-orange-800'
      case 'zset':
        return 'bg-red-100 text-red-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatTTL = (ttl: number) => {
    if (ttl === -1) return 'No expiry'
    if (ttl === -2) return 'Expired'
    if (ttl < 60) return `${ttl}s`
    if (ttl < 3600) return `${Math.floor(ttl / 60)}m`
    if (ttl < 86400) return `${Math.floor(ttl / 3600)}h`
    return `${Math.floor(ttl / 86400)}d`
  }

  const filteredKeys = keys?.filter(key => 
    dataType === 'all' || key.type === dataType
  ) || []

  return (
    <motion.div
      className={cn("space-y-6", className)}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Redis Data Explorer</h1>
          <p className="text-muted-foreground">
            Explore and visualize your Redis data structures
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleSearch}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          <Button variant="outline" size="sm">
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
        </div>
      </div>

      {/* Search and Filters */}
      <Card>
        <CardContent className="p-4">
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search keys (use * for wildcards)"
                  value={searchPattern}
                  onChange={(e) => setSearchPattern(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
                  className="pl-10"
                />
              </div>
            </div>
            <Select value={dataType} onValueChange={setDataType}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="Data type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="string">String</SelectItem>
                <SelectItem value="hash">Hash</SelectItem>
                <SelectItem value="list">List</SelectItem>
                <SelectItem value="set">Set</SelectItem>
                <SelectItem value="zset">Sorted Set</SelectItem>
                <SelectItem value="stream">Stream</SelectItem>
              </SelectContent>
            </Select>
            <Select value={limit.toString()} onValueChange={(value) => setLimit(parseInt(value))}>
              <SelectTrigger className="w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="50">50</SelectItem>
                <SelectItem value="100">100</SelectItem>
                <SelectItem value="500">500</SelectItem>
                <SelectItem value="1000">1000</SelectItem>
              </SelectContent>
            </Select>
            <Button onClick={handleSearch}>
              <Search className="h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Keys List */}
        <div className="lg:col-span-1">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Database className="h-5 w-5" />
                Keys ({filteredKeys.length})
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {isLoading ? (
                <div className="p-4 text-center text-muted-foreground">
                  Loading keys...
                </div>
              ) : error ? (
                <div className="p-4 text-center text-red-500">
                  Error: {error}
                </div>
              ) : filteredKeys.length === 0 ? (
                <div className="p-4 text-center text-muted-foreground">
                  No keys found
                </div>
              ) : (
                <div className="max-h-96 overflow-y-auto">
                  {filteredKeys.map((key, index) => (
                    <motion.div
                      key={key.key}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.05 }}
                      className={cn(
                        "p-3 border-b cursor-pointer hover:bg-muted/50 transition-colors",
                        selectedKey === key.key && "bg-muted"
                      )}
                      onClick={() => handleKeySelect(key.key)}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2 min-w-0 flex-1">
                          {getTypeIcon(key.type)}
                          <span className="font-mono text-sm truncate">
                            {key.key}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <Badge className={getTypeColor(key.type)}>
                            {key.type}
                          </Badge>
                          {key.ttl > 0 && (
                            <Badge variant="outline" className="text-xs">
                              <Clock className="h-3 w-3 mr-1" />
                              {formatTTL(key.ttl)}
                            </Badge>
                          )}
                        </div>
                      </div>
                      {(key.length || key.cardinality || key.field_count) && (
                        <div className="mt-1 text-xs text-muted-foreground">
                          {key.length && `Length: ${key.length}`}
                          {key.cardinality && `Size: ${key.cardinality}`}
                          {key.field_count && `Fields: ${key.field_count}`}
                        </div>
                      )}
                    </motion.div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Key Details */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Eye className="h-5 w-5" />
                {selectedKey ? `Key Details: ${selectedKey}` : 'Select a key to view details'}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {selectedKey && keyDetails ? (
                <Tabs value={activeTab} onValueChange={setActiveTab}>
                  <TabsList>
                    <TabsTrigger value="overview">Overview</TabsTrigger>
                    <TabsTrigger value="data">Data</TabsTrigger>
                    <TabsTrigger value="actions">Actions</TabsTrigger>
                  </TabsList>
                  
                  <TabsContent value="overview" className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <label className="text-sm font-medium">Key</label>
                        <p className="font-mono text-sm bg-muted p-2 rounded">
                          {keyDetails.key}
                        </p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">Type</label>
                        <div className="mt-1">
                          <Badge className={getTypeColor(keyDetails.type)}>
                            {getTypeIcon(keyDetails.type)}
                            {keyDetails.type}
                          </Badge>
                        </div>
                      </div>
                      <div>
                        <label className="text-sm font-medium">TTL</label>
                        <p className="text-sm">{formatTTL(keyDetails.ttl)}</p>
                      </div>
                      <div>
                        <label className="text-sm font-medium">Memory Usage</label>
                        <p className="text-sm">
                          {keyDetails.memory_usage ? 
                            `${(keyDetails.memory_usage / 1024).toFixed(2)} KB` : 
                            'N/A'
                          }
                        </p>
                      </div>
                    </div>
                  </TabsContent>
                  
                  <TabsContent value="data">
                    <div className="text-center text-muted-foreground">
                      Data visualization will be implemented here
                    </div>
                  </TabsContent>
                  
                  <TabsContent value="actions" className="space-y-4">
                    <div className="flex gap-2">
                      <Button variant="outline" size="sm">
                        <Eye className="h-4 w-4 mr-2" />
                        View Raw
                      </Button>
                      <Button variant="outline" size="sm">
                        <Download className="h-4 w-4 mr-2" />
                        Export
                      </Button>
                      <Button variant="destructive" size="sm">
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </Button>
                    </div>
                  </TabsContent>
                </Tabs>
              ) : (
                <div className="text-center text-muted-foreground py-8">
                  Select a key from the list to view its details
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </motion.div>
  )
}
