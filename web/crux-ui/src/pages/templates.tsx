import { Layout } from '@app/components/layout'
import { BreadcrumbLink } from '@app/components/shared/breadcrumb'
import PageHeading from '@app/components/shared/page-heading'
import ApplyTemplateCard from '@app/components/templates/apply-template-card'
import TemplateCard from '@app/components/templates/template-card'
import DyoButton from '@app/elements/dyo-button'
import DyoWrap from '@app/elements/dyo-wrap'
import { Template } from '@app/models/template'
import { productUrl, ROUTE_TEMPLATES } from '@app/routes'
import { withContextAuthorization } from '@app/utils'
import { cruxFromContext } from '@server/crux/crux'
import { NextPageContext } from 'next'
import useTranslation from 'next-translate/useTranslation'
import { useRouter } from 'next/router'
import { useRef, useState } from 'react'

interface TemplatesPageProps {
  templates: Template[]
}

const TemplatesPage = (props: TemplatesPageProps) => {
  const { templates } = props

  const router = useRouter()

  const { t } = useTranslation('products')

  const [applying, setApplying] = useState<Template | null>(null)
  const submitRef = useRef<() => Promise<any>>()

  const pageLink: BreadcrumbLink = {
    name: t('common:templates'),
    url: ROUTE_TEMPLATES,
  }

  const onCreate = (template: Template) => {
    setApplying(template)
  }

  const onTemplateApplied = productId => {
    setApplying(null)
    router.push(productUrl(productId))
  }

  return (
    <Layout title={t('common:templates')}>
      <PageHeading pageLink={pageLink}>
        {applying && (
          <>
            <DyoButton className="ml-auto px-4" secondary onClick={() => setApplying(null)}>
              {t('common:discard')}
            </DyoButton>

            <DyoButton className="px-4 ml-4" onClick={() => submitRef.current()}>
              {t('common:add')}
            </DyoButton>
          </>
        )}
      </PageHeading>

      {applying && (
        <ApplyTemplateCard template={applying} onTemplateApplied={onTemplateApplied} submitRef={submitRef} />
      )}

      <DyoWrap>
        {templates.map(it => (
          <TemplateCard key={it.id} template={it} onAddClick={() => onCreate(it)} />
        ))}
      </DyoWrap>
    </Layout>
  )
}
export default TemplatesPage

const getPageServerSideProps = async (context: NextPageContext) => ({
  props: {
    templates: await cruxFromContext(context).templates.getAll(),
  },
})

export const getServerSideProps = withContextAuthorization(getPageServerSideProps)
