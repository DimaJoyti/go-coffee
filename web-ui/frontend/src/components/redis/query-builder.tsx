'use client'

import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Play, 
  Save, 
  Copy, 
  History, 
  Lightbulb,
  Code,
  CheckCircle,
  XCircle,
  Terminal,
  Zap
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { useRedisQuery } from '@/hooks/use-redis'
import { cn } from '@/lib/utils'

interface QueryBuilderProps {
  className?: string
}

interface QueryTemplate {
  name: string
  description: string
  operation: string
  key?: string
  field?: string
  value?: any
  args?: any[]
  example: string
}

export function QueryBuilder({ className }: QueryBuilderProps) {
  const [operation, setOperation] = useState('GET')
  const [key, setKey] = useState('')
  const [field, setField] = useState('')
  const [value, setValue] = useState('')
  const [args, setArgs] = useState('')
  const [preview, setPreview] = useState(true)
  const [activeTab, setActiveTab] = useState('builder')
  
  const {
    buildQuery,
    validateQuery,
    executeQuery,
    templates,
    suggestions,
    queryResult,
    isLoading,
    error,
    getTemplates,
    getSuggestions
  } = useRedisQuery()

  useEffect(() => {
    getTemplates(operation)
    getSuggestions(operation)
  }, [operation, getTemplates, getSuggestions])

  const handleBuildQuery = async () => {
    const queryRequest = {
      operation,
      key,
      field: field || undefined,
      value: value || undefined,
      args: args ? args.split(',').map(arg => arg.trim()) : [],
      preview
    }

    if (preview) {
      await buildQuery(queryRequest)
    } else {
      await executeQuery(queryRequest)
    }
  }

  const handleValidateQuery = async () => {
    const queryRequest = {
      operation,
      key,
      field: field || undefined,
      value: value || undefined,
      args: args ? args.split(',').map(arg => arg.trim()) : [],
      preview: true
    }

    await validateQuery(queryRequest)
  }

  const handleTemplateSelect = (template: QueryTemplate) => {
    setOperation(template.operation)
    setKey(template.key || '')
    setField(template.field || '')
    setValue(template.value || '')
    setArgs(template.args ? template.args.join(', ') : '')
  }

  const getOperationColor = (op: string) => {
    const readOps = ['GET', 'HGET', 'LRANGE', 'SMEMBERS', 'ZRANGE', 'EXISTS', 'TTL']
    const writeOps = ['SET', 'HSET', 'LPUSH', 'RPUSH', 'SADD', 'ZADD']
    const deleteOps = ['DEL', 'HDEL', 'LPOP', 'RPOP', 'SREM', 'ZREM']
    
    if (readOps.includes(op)) return 'bg-blue-100 text-blue-800'
    if (writeOps.includes(op)) return 'bg-green-100 text-green-800'
    if (deleteOps.includes(op)) return 'bg-red-100 text-red-800'
    return 'bg-gray-100 text-gray-800'
  }

  const operations = [
    { value: 'GET', label: 'GET - Get string value' },
    { value: 'SET', label: 'SET - Set string value' },
    { value: 'HGET', label: 'HGET - Get hash field' },
    { value: 'HSET', label: 'HSET - Set hash field' },
    { value: 'LPUSH', label: 'LPUSH - Push to list head' },
    { value: 'RPUSH', label: 'RPUSH - Push to list tail' },
    { value: 'SADD', label: 'SADD - Add to set' },
    { value: 'ZADD', label: 'ZADD - Add to sorted set' },
    { value: 'DEL', label: 'DEL - Delete key' },
    { value: 'EXISTS', label: 'EXISTS - Check if key exists' },
    { value: 'TTL', label: 'TTL - Get time to live' },
    { value: 'EXPIRE', label: 'EXPIRE - Set expiration' },
  ]

  const needsField = ['HGET', 'HSET', 'HDEL'].includes(operation)
  const needsValue = ['SET', 'HSET', 'LPUSH', 'RPUSH', 'SADD', 'ZADD'].includes(operation)
  const needsArgs = ['ZADD', 'EXPIRE', 'LRANGE', 'ZRANGE'].includes(operation)

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
          <h1 className="text-2xl font-bold">Redis Query Builder</h1>
          <p className="text-muted-foreground">
            Build and execute Redis commands visually
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm">
            <History className="h-4 w-4 mr-2" />
            History
          </Button>
          <Button variant="outline" size="sm">
            <Save className="h-4 w-4 mr-2" />
            Save Query
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Query Builder */}
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Code className="h-5 w-5" />
                Query Builder
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <Tabs value={activeTab} onValueChange={setActiveTab}>
                <TabsList>
                  <TabsTrigger value="builder">Visual Builder</TabsTrigger>
                  <TabsTrigger value="raw">Raw Command</TabsTrigger>
                </TabsList>
                
                <TabsContent value="builder" className="space-y-4">
                  {/* Operation Selection */}
                  <div>
                    <Label htmlFor="operation">Operation</Label>
                    <Select value={operation} onValueChange={setOperation}>
                      <SelectTrigger>
                        <SelectValue placeholder="Select operation" />
                      </SelectTrigger>
                      <SelectContent>
                        {operations.map((op) => (
                          <SelectItem key={op.value} value={op.value}>
                            <div className="flex items-center gap-2">
                              <Badge className={getOperationColor(op.value)}>
                                {op.value}
                              </Badge>
                              <span className="text-sm">{op.label.split(' - ')[1]}</span>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  {/* Key Input */}
                  <div>
                    <Label htmlFor="key">Key</Label>
                    <Input
                      id="key"
                      placeholder="Enter Redis key"
                      value={key}
                      onChange={(e) => setKey(e.target.value)}
                    />
                  </div>

                  {/* Field Input (for hash operations) */}
                  {needsField && (
                    <div>
                      <Label htmlFor="field">Field</Label>
                      <Input
                        id="field"
                        placeholder="Enter hash field"
                        value={field}
                        onChange={(e) => setField(e.target.value)}
                      />
                    </div>
                  )}

                  {/* Value Input */}
                  {needsValue && (
                    <div>
                      <Label htmlFor="value">Value</Label>
                      <Textarea
                        id="value"
                        placeholder="Enter value"
                        value={value}
                        onChange={(e) => setValue(e.target.value)}
                        rows={3}
                      />
                    </div>
                  )}

                  {/* Additional Arguments */}
                  {needsArgs && (
                    <div>
                      <Label htmlFor="args">
                        Additional Arguments
                        {operation === 'ZADD' && ' (score)'}
                        {operation === 'EXPIRE' && ' (seconds)'}
                        {operation === 'LRANGE' && ' (start, stop)'}
                      </Label>
                      <Input
                        id="args"
                        placeholder={
                          operation === 'ZADD' ? 'Enter score (e.g., 100.5)' :
                          operation === 'EXPIRE' ? 'Enter seconds (e.g., 3600)' :
                          'Enter arguments separated by commas'
                        }
                        value={args}
                        onChange={(e) => setArgs(e.target.value)}
                      />
                    </div>
                  )}

                  {/* Preview Mode Toggle */}
                  <div className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      id="preview"
                      checked={preview}
                      onChange={(e) => setPreview(e.target.checked)}
                      className="rounded"
                    />
                    <Label htmlFor="preview">Preview mode (don't execute)</Label>
                  </div>

                  {/* Action Buttons */}
                  <div className="flex gap-2">
                    <Button onClick={handleValidateQuery} variant="outline">
                      <CheckCircle className="h-4 w-4 mr-2" />
                      Validate
                    </Button>
                    <Button onClick={handleBuildQuery} disabled={isLoading}>
                      {preview ? (
                        <>
                          <Code className="h-4 w-4 mr-2" />
                          Build Query
                        </>
                      ) : (
                        <>
                          <Play className="h-4 w-4 mr-2" />
                          Execute
                        </>
                      )}
                    </Button>
                    <Button variant="outline">
                      <Copy className="h-4 w-4 mr-2" />
                      Copy
                    </Button>
                  </div>
                </TabsContent>
                
                <TabsContent value="raw">
                  <div className="space-y-4">
                    <div>
                      <Label htmlFor="raw-command">Raw Redis Command</Label>
                      <Textarea
                        id="raw-command"
                        placeholder="Enter Redis command (e.g., GET user:123)"
                        rows={4}
                        className="font-mono"
                      />
                    </div>
                    <Button>
                      <Terminal className="h-4 w-4 mr-2" />
                      Execute Raw Command
                    </Button>
                  </div>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>

          {/* Query Result */}
          {queryResult && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  {queryResult.success ? (
                    <CheckCircle className="h-5 w-5 text-green-500" />
                  ) : (
                    <XCircle className="h-5 w-5 text-red-500" />
                  )}
                  Query Result
                </CardTitle>
              </CardHeader>
              <CardContent>
                {queryResult.success ? (
                  <div className="space-y-2">
                    {queryResult.redis_cmd && (
                      <div>
                        <Label>Generated Command:</Label>
                        <pre className="bg-muted p-2 rounded text-sm font-mono">
                          {queryResult.redis_cmd}
                        </pre>
                      </div>
                    )}
                    {queryResult.result !== undefined && (
                      <div>
                        <Label>Result:</Label>
                        <pre className="bg-muted p-2 rounded text-sm">
                          {JSON.stringify(queryResult.result, null, 2)}
                        </pre>
                      </div>
                    )}
                    {queryResult.preview && (
                      <div>
                        <Label>Preview:</Label>
                        <p className="text-sm text-muted-foreground">
                          {queryResult.preview}
                        </p>
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="text-red-500">
                    <p>Error: {error}</p>
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>

        {/* Templates and Suggestions */}
        <div className="space-y-6">
          {/* Templates */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Lightbulb className="h-5 w-5" />
                Templates
              </CardTitle>
            </CardHeader>
            <CardContent>
              {templates && templates.length > 0 ? (
                <div className="space-y-2">
                  {templates.map((template, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, x: 20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.1 }}
                      className="p-3 border rounded cursor-pointer hover:bg-muted/50 transition-colors"
                      onClick={() => handleTemplateSelect(template)}
                    >
                      <div className="font-medium text-sm">{template.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {template.description}
                      </div>
                      <div className="text-xs font-mono mt-1 text-blue-600">
                        {template.example}
                      </div>
                    </motion.div>
                  ))}
                </div>
              ) : (
                <div className="text-sm text-muted-foreground">
                  No templates available for {operation}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Suggestions */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5" />
                Suggestions
              </CardTitle>
            </CardHeader>
            <CardContent>
              {suggestions && suggestions.length > 0 ? (
                <div className="space-y-1">
                  {suggestions.map((suggestion, index) => (
                    <div
                      key={index}
                      className="text-xs font-mono p-2 bg-muted rounded cursor-pointer hover:bg-muted/80"
                    >
                      {suggestion}
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-sm text-muted-foreground">
                  No suggestions available
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </motion.div>
  )
}
