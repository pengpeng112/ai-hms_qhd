import { BarChart, Bar, XAxis, Tooltip, ResponsiveContainer } from 'recharts'

interface ChartCardProps {
  data: { name: string; value: number }[]
  color?: string
}

export default function ChartCard({ data, color = '#3b82f6' }: ChartCardProps) {
  return (
    <div className="h-full min-h-[200px] w-full">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data}>
          <XAxis dataKey="name" fontSize={10} tickLine={false} axisLine={false} tick={{ fill: '#9ca3af' }} />
          <Tooltip
            contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}
            cursor={{ fill: '#f3f4f6' }}
          />
          <Bar dataKey="value" fill={color} radius={[4, 4, 0, 0]} barSize={24} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}
