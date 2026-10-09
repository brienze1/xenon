import Markdown from 'react-markdown'
import styles from './InfoPanel.module.css'

interface MarkdownViewProps {
  text: string
}

export default function MarkdownView({ text }: MarkdownViewProps) {
  return (
    <div className={styles.markdown}>
      <Markdown>{text}</Markdown>
    </div>
  )
}
