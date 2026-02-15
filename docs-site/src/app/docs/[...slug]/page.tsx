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

function extractFaqPairs(markdown: string): { question: string; answer: string }[] {
  const faqHeadingIndex = markdown.search(/^## Frequently Asked Questions\s*$/m);
  if (faqHeadingIndex === -1) return [];

  const faqSection = markdown.slice(faqHeadingIndex);
  const pairs: { question: string; answer: string }[] = [];
  const questionRegex = /^### (.+)$/gm;
  let match: RegExpExecArray | null;
  const matches: { question: string; start: number }[] = [];

  while ((match = questionRegex.exec(faqSection)) !== null) {
    matches.push({ question: match[1].trim(), start: match.index + match[0].length });
  }

  for (let i = 0; i < matches.length; i++) {
    const end = i + 1 < matches.length ? faqSection.indexOf(`### ${matches[i + 1].question}`, matches[i].start) : faqSection.length;
    const answer = faqSection.slice(matches[i].start, end).trim();
    if (answer) {
      pairs.push({ question: matches[i].question, answer });
    }
  }

  return pairs;
}

export default async function DocPage({ params }: PageProps) {
  const resolvedParams = await params;
  const slugPath = resolvedParams.slug.join('/').toLowerCase();
  const doc = getDocContent(slugPath);
  if (!doc) {
    notFound();
  }

  const html = markdownToHtml(doc.content, slugPath);
  const faqPairs = extractFaqPairs(doc.content);
  const faqJsonLd = faqPairs.length > 0 ? {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: faqPairs.map((pair) => ({
      '@type': 'Question',
      name: pair.question,
      acceptedAnswer: {
        '@type': 'Answer',
        text: pair.answer,
      },
    })),
  } : null;

  return (
    <div>
      {faqJsonLd && (
        <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }} />
      )}
      <h1 className="text-3xl font-bold text-white mb-6">{doc.title}</h1>
      <MarkdownRenderer html={html} />
    </div>
  );
}
