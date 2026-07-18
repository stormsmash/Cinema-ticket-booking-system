import fourKingsPoster from '@/assets/posters/thai/4-kings-2.webp'
import badGeniusPoster from '@/assets/posters/thai/bad-genius.webp'
import deathWhispererPoster from '@/assets/posters/thai/death-whisperer-2.webp'
import lahnMahPoster from '@/assets/posters/thai/lahn-mah.webp'
import myBooPoster from '@/assets/posters/thai/my-boo.webp'
import oneDayPoster from '@/assets/posters/thai/one-day.webp'
import paradiseOfThornsPoster from '@/assets/posters/thai/paradise-of-thorns.webp'
import peeMakPoster from '@/assets/posters/thai/pee-mak.webp'
import theMediumPoster from '@/assets/posters/thai/the-medium.webp'
import theUndertakerPoster from '@/assets/posters/thai/the-undertaker.webp'

export interface MoviePresentation {
  title: string
  englishTitle: string
  poster: string
  genres: string[]
  certificate: string
  language: string
  year: number
  description: string
  sourceUrl: string
  featured?: boolean
}

export const movieCatalog: MoviePresentation[] = [
  {
    title: 'หลานม่า',
    englishTitle: 'How to Make Millions Before Grandma Dies',
    poster: lahnMahPoster,
    genres: ['ดราม่า', 'ครอบครัว'],
    certificate: 'น 13+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2024,
    description:
      'เอ็มอาสาดูแลอาม่าที่ป่วยเพราะหวังมรดก แต่ช่วงเวลาที่ได้อยู่ด้วยกันทำให้เขาค่อย ๆ เข้าใจความหมายของครอบครัว',
    sourceUrl: 'https://www.themoviedb.org/movie/1103621',
    featured: true,
  },
  {
    title: 'ธี่หยด 2',
    englishTitle: 'Death Whisperer 2',
    poster: deathWhispererPoster,
    genres: ['สยองขวัญ', 'ระทึกขวัญ'],
    certificate: 'ฉ 18+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2024,
    description:
      'ยักษ์ยังคงตามล่าผีชุดดำที่พรากน้องสาวไป แม้ครอบครัวจะขอให้หยุด เพราะภัยร้ายอาจย้อนกลับมาหาทุกคนอีกครั้ง',
    sourceUrl: 'https://www.themoviedb.org/movie/1247019',
  },
  {
    title: 'วิมานหนาม',
    englishTitle: 'The Paradise of Thorns',
    poster: paradiseOfThornsPoster,
    genres: ['ดราม่า', 'ระทึกขวัญ'],
    certificate: 'น 15+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2024,
    description:
      'เมื่อคนรักเสียชีวิตกะทันหัน ทองคำต้องต่อสู้เพื่อบ้านและสวนทุเรียนที่ทั้งคู่ช่วยกันสร้าง แต่กฎหมายกลับไม่ยอมรับสิทธิของเขา',
    sourceUrl: 'https://www.themoviedb.org/movie/1290206',
  },
  {
    title: 'อนงค์',
    englishTitle: 'My Boo',
    poster: myBooPoster,
    genres: ['โรแมนติก', 'คอมเมดี้'],
    certificate: 'น 13+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2024,
    description:
      'โจได้รับมรดกเป็นบ้านเก่าพร้อมอนงค์ ผีเจ้าของบ้าน ทั้งสองจึงเปิดบ้านผีสิงหาเงินและค่อย ๆ เปลี่ยนความวุ่นวายเป็นความผูกพัน',
    sourceUrl: 'https://www.themoviedb.org/movie/1257388',
  },
  {
    title: 'สัปเหร่อ',
    englishTitle: 'The Undertaker',
    poster: theUndertakerPoster,
    genres: ['คอมเมดี้', 'สยองขวัญ'],
    certificate: 'น 15+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2023,
    description:
      'หนุ่มที่กลัวผีต้องมาช่วยงานสัปเหร่อ ขณะที่อีกคนพยายามหาทางพบคนรักที่จากไป เรื่องความตายจึงพาทุกคนกลับมาเรียนรู้การใช้ชีวิต',
    sourceUrl: 'https://www.themoviedb.org/movie/1113119',
  },
  {
    title: '4 Kings II',
    englishTitle: '4 Kings II',
    poster: fourKingsPoster,
    genres: ['แอ็กชัน', 'ดราม่า'],
    certificate: 'ฉ 18+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2023,
    description:
      'ความขัดแย้งระหว่างสองสถาบันอาชีวะลุกลามเมื่อกลุ่มเด็กบ้านเข้ามาเกี่ยวข้อง การเอาคืนจึงนำไปสู่ความสูญเสียที่ไม่มีใครควบคุมได้',
    sourceUrl: 'https://www.themoviedb.org/movie/968232',
  },
  {
    title: 'ร่างทรง',
    englishTitle: 'The Medium',
    poster: theMediumPoster,
    genres: ['สยองขวัญ', 'ลึกลับ'],
    certificate: 'ฉ 18+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2021,
    description:
      'ทีมสารคดีติดตามครอบครัวร่างทรงในภาคอีสาน เมื่ออาการผิดปกติของหญิงสาวในบ้านเริ่มรุนแรง ความเชื่อเดิมจึงถูกท้าทาย',
    sourceUrl: 'https://www.themoviedb.org/movie/745881',
  },
  {
    title: 'ฉลาดเกมส์โกง',
    englishTitle: 'Bad Genius',
    poster: badGeniusPoster,
    genres: ['ระทึกขวัญ', 'ดราม่า'],
    certificate: 'น 13+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2017,
    description:
      'ลินเปลี่ยนการช่วยเพื่อนลอกข้อสอบให้เป็นธุรกิจ ก่อนแผนโกงครั้งใหญ่จะพาเธอข้ามประเทศและเดิมพันอนาคตของทุกคนในทีม',
    sourceUrl: 'https://www.themoviedb.org/movie/455714',
  },
  {
    title: 'พี่มาก..พระโขนง',
    englishTitle: 'Pee Mak',
    poster: peeMakPoster,
    genres: ['คอมเมดี้', 'สยองขวัญ'],
    certificate: 'น 13+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2013,
    description:
      'มากกลับจากสงครามมาหานาคและลูกชาย โดยไม่รู้ข่าวลือที่เพื่อนทั้งสี่ได้ยินมา ภารกิจบอกความจริงจึงทั้งน่ากลัวและชวนหัว',
    sourceUrl: 'https://www.themoviedb.org/movie/184219',
  },
  {
    title: 'แฟนเดย์..แฟนกันแค่วันเดียว',
    englishTitle: 'One Day',
    poster: oneDayPoster,
    genres: ['โรแมนติก', 'ดราม่า'],
    certificate: 'น 13+',
    language: 'ภาษาไทย · คำบรรยายอังกฤษ',
    year: 2016,
    description:
      'เด่นชัยขอพรให้ได้เป็นแฟนกับนุ้ยเพียงวันเดียว เมื่อเธอสูญเสียความทรงจำชั่วคราวระหว่างทริปบริษัท เขาจึงต้องตัดสินใจว่าจะใช้โอกาสนั้นอย่างไร',
    sourceUrl: 'https://www.themoviedb.org/movie/420541',
  },
]

const catalogByTitle = new Map(movieCatalog.map((movie) => [movie.title, movie]))

export function getMoviePresentation(title: string): MoviePresentation {
  return (
    catalogByTitle.get(title) ?? {
      title,
      englishTitle: title,
      poster: lahnMahPoster,
      genres: ['ภาพยนตร์'],
      certificate: 'ท',
      language: 'ภาษาไทย',
      year: new Date().getFullYear(),
      description: 'เลือกภาพยนตร์ รอบฉาย และที่นั่งที่ต้องการเพื่อดำเนินการจอง',
      sourceUrl: '',
    }
  )
}
