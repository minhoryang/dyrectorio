import DyoButton from '@app/elements/dyo-button'
import { DyoCard } from '@app/elements/dyo-card'
import { DyoHeading } from '@app/elements/dyo-heading'
import { Template } from '@app/models/template'
import useTranslation from 'next-translate/useTranslation'
import Image from 'next/image'

export interface TemplateCardProps {
  template: Template
  onAddClick: VoidFunction
}

const TemplateCard = (props: TemplateCardProps) => {
  const { template: propsTemplate, onAddClick } = props
  const { name, description } = propsTemplate

  const { t } = useTranslation('templates')

  return (
    <DyoCard className="p-6 flex flex-col flex-grow w-full">
      <div className="flex flex-col w-full">
        <div className="flex flex-row">
          <Image src="/default_template.svg" width={100} height={100} />

          <div className="flex flex-col flex-grow">
            <DyoHeading element="h5" className="text-lg text-bright ml-4">
              {name}
            </DyoHeading>
          </div>

          <DyoButton className="ml-auto px-4" onClick={onAddClick}>
            {t('common:add')}
          </DyoButton>
        </div>

        <p className="text-md text-bright mt-4 line-clamp-2 break-words">{description}</p>
      </div>
    </DyoCard>
  )
}

export default TemplateCard
