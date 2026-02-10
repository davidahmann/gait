import { notFound } from 'next/navigation';
import type { Metadata } from 'next';
import MarkdownRenderer from '@/components/MarkdownRenderer';
import { getAllDocSlugs, getDocContent } from '@/lib/docs';
import { markdownToHtml } from '@/lib/markdown';
import { canonicalUrl } from '@/lib/site';

interface PageProps {
  params: Promise<{ slug: string[] }>;
}

export function generateStaticParams() {
  return getAllDocSlugs().map((slug) => ({
    slug: slug.split('/'),
  }));
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  const resolvedParams = await params;
  const slugPath = resolvedParams.slug.join('/').toLowerCase();
  const doc = getDocContent(slugPath);

  if (!doc) {
    return {
      title: 'Not Found | Gait Docs',
      description: 'Requested documentation page was not found.',
    };
  }

  const canonical = canonicalUrl(`/docs/${slugPath}/`);

  return {
    title: `${doc.title} | Gait Docs`,
    description: doc.description || `Gait documentation page: ${doc.title}`,
    alternates: { canonical },
    openGraph: {
      title: `${doc.title} | Gait Docs`,
      description: doc.description || `Gait documentation page: ${doc.title}`,
      url: canonical,
      type: 'article',
    },
  };
}

export default async function DocPage({ params }: PageProps) {
  const resolvedParams = await params;
  const slugPath = resolvedParams.slug.join('/').toLowerCase();
  const doc = getDocContent(slugPath);
  if (!doc) {
    notFound();
  }

  const html = markdownToHtml(doc.content, slugPath);

  return (
    <div>
      <h1 className="text-3xl font-bold text-white mb-6">{doc.title}</h1>
      <MarkdownRenderer html={html} />
    </div>
  );
}
