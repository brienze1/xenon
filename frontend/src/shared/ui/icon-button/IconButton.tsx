import type { ComponentProps } from 'react'
import { cn } from '@/shared/lib/cn'
import styles from './IconButton.module.css'

interface IconButtonProps extends Omit<ComponentProps<'button'>, 'aria-label'> {
  label: string
}

export function IconButton({ label, className, type = 'button', title, ...props }: IconButtonProps) {
  return (
    <button
      type={type}
      aria-label={label}
      title={title ?? label}
      className={cn(styles.iconButton, className)}
      {...props}
    />
  )
}
