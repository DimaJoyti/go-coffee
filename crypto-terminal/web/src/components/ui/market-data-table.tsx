import * as React from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./table"
import { Card, CardContent, CardHeader, CardTitle } from "./card"
import { Badge } from "./badge"
import { Button } from "./button"
import { Input } from "./input"
import { cn, formatCurrency, formatPercentage, getChangeColor } from "../../lib/utils"
import { Search, ArrowUpDown, Star, TrendingUp, TrendingDown } from "lucide-react"

interface MarketDataItem {
  id: string
  symbol: string
  name: string
  price: number
  change24h: number
  changePercent24h: number
  volume24h: number
  marketCap: number
  rank?: number
  logo?: string
  isFavorite?: boolean
}

interface MarketDataTableProps {
  data: MarketDataItem[]
  title?: string
  currency?: string
  loading?: boolean
  onSort?: (column: string, direction: "asc" | "desc") => void
  onFavorite?: (symbol: string) => void
  onRowClick?: (item: MarketDataItem) => void
  showSearch?: boolean
  showRank?: boolean
  showVolume?: boolean
  showMarketCap?: boolean
  className?: string
}

const MarketDataTable: React.FC<MarketDataTableProps> = ({
  data,
  title = "Market Data",
  currency = "USD",
  loading = false,
  onSort,
  onFavorite,
  onRowClick,
  showSearch = true,
  showRank = true,
  showVolume = true,
  showMarketCap = true,
  className,
}) => {
  const [searchTerm, setSearchTerm] = React.useState("")
  const [sortColumn, setSortColumn] = React.useState<string>("")
  const [sortDirection, setSortDirection] = React.useState<"asc" | "desc">("desc")

  const filteredData = React.useMemo(() => {
    return data.filter(item =>
      item.symbol.toLowerCase().includes(searchTerm.toLowerCase()) ||
      item.name.toLowerCase().includes(searchTerm.toLowerCase())
    )
  }, [data, searchTerm])

  const handleSort = (column: string) => {
    const newDirection = sortColumn === column && sortDirection === "desc" ? "asc" : "desc"
    setSortColumn(column)
    setSortDirection(newDirection)
    onSort?.(column, newDirection)
  }

  const SortableHeader = ({ column, children }: { column: string; children: React.ReactNode }) => (
    <TableHead 
      className="cursor-pointer hover:bg-muted/50 transition-colors"
      onClick={() => handleSort(column)}
    >
      <div className="flex items-center space-x-1">
        <span>{children}</span>
        <ArrowUpDown className="h-3 w-3 opacity-50" />
      </div>
    </TableHead>
  )

  const LoadingRow = () => (
    <TableRow>
      {showRank && <TableCell><div className="h-4 bg-muted rounded w-8"></div></TableCell>}
      <TableCell><div className="h-4 bg-muted rounded w-16"></div></TableCell>
      <TableCell><div className="h-4 bg-muted rounded w-20"></div></TableCell>
      <TableCell><div className="h-4 bg-muted rounded w-16"></div></TableCell>
      <TableCell><div className="h-4 bg-muted rounded w-12"></div></TableCell>
      {showVolume && <TableCell><div className="h-4 bg-muted rounded w-20"></div></TableCell>}
      {showMarketCap && <TableCell><div className="h-4 bg-muted rounded w-20"></div></TableCell>}
      <TableCell><div className="h-4 bg-muted rounded w-8"></div></TableCell>
    </TableRow>
  )

  return (
    <Card className={cn("w-full", className)}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>{title}</CardTitle>
          {showSearch && (
            <div className="relative w-64">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search cryptocurrencies..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-8"
              />
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent>
        <div className="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                {showRank && <SortableHeader column="rank">#</SortableHeader>}
                <SortableHeader column="symbol">Asset</SortableHeader>
                <SortableHeader column="name">Name</SortableHeader>
                <SortableHeader column="price">Price</SortableHeader>
                <SortableHeader column="changePercent24h">24h %</SortableHeader>
                {showVolume && <SortableHeader column="volume24h">Volume (24h)</SortableHeader>}
                {showMarketCap && <SortableHeader column="marketCap">Market Cap</SortableHeader>}
                <TableHead className="w-12">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 10 }).map((_, index) => (
                  <LoadingRow key={index} />
                ))
              ) : (
                filteredData.map((item) => (
                  <TableRow
                    key={item.id}
                    className={cn(
                      "cursor-pointer hover:bg-muted/50 transition-colors",
                      onRowClick && "hover:bg-accent"
                    )}
                    onClick={() => onRowClick?.(item)}
                  >
                    {showRank && (
                      <TableCell className="font-medium text-muted-foreground">
                        {item.rank || "-"}
                      </TableCell>
                    )}
                    <TableCell>
                      <div className="flex items-center space-x-2">
                        {item.logo && (
                          <img 
                            src={item.logo} 
                            alt={item.symbol} 
                            className="w-6 h-6 rounded-full"
                          />
                        )}
                        <span className="font-medium">{item.symbol}</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {item.name}
                    </TableCell>
                    <TableCell className="font-mono">
                      {formatCurrency(item.price, currency)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-1">
                        {item.changePercent24h > 0 ? (
                          <TrendingUp className="h-3 w-3 text-bull" />
                        ) : (
                          <TrendingDown className="h-3 w-3 text-bear" />
                        )}
                        <span className={cn(
                          "font-medium",
                          getChangeColor(item.changePercent24h)
                        )}>
                          {formatPercentage(item.changePercent24h / 100)}
                        </span>
                      </div>
                    </TableCell>
                    {showVolume && (
                      <TableCell className="font-mono text-muted-foreground">
                        {formatCurrency(item.volume24h, currency, 0)}
                      </TableCell>
                    )}
                    {showMarketCap && (
                      <TableCell className="font-mono text-muted-foreground">
                        {formatCurrency(item.marketCap, currency, 0)}
                      </TableCell>
                    )}
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={(e) => {
                          e.stopPropagation()
                          onFavorite?.(item.symbol)
                        }}
                        className="h-8 w-8"
                      >
                        <Star 
                          className={cn(
                            "h-4 w-4",
                            item.isFavorite ? "fill-yellow-400 text-yellow-400" : "text-muted-foreground"
                          )} 
                        />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
        {!loading && filteredData.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            No cryptocurrencies found matching your search.
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export { MarketDataTable, type MarketDataItem }
