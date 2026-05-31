interface PlaceholderCardProps {
  label: string
}

export default function PlaceholderCard({ label }: PlaceholderCardProps) {
  return (
    <div className="flex items-center justify-center h-full text-foreground-muted text-sm">
      {label}（待接口）
    </div>
  )
}
