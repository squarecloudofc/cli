import { promises as fs } from "fs";
import { join, resolve } from "path";

const DEFAULT_LOCALE = "pt";
const LOCALES = ["en", "pt", "es", "zh"];
const LOCALES_DIR = resolve("i18n/resources");

/**
 * @param {string} filePath
 */
const readJson = async (filePath) => {
  try {
    const content = await fs.readFile(filePath, "utf-8");
    return JSON.parse(content);
  } catch (err) {
    console.warn(`Aviso: falha ao ler ${filePath}: ${err.message}`);
    return null;
  }
};

/**
 * @param {string} filePath
 * @param {Record<string, unknown>} obj
 */
async function writeJson(filePath, obj) {
  const content = JSON.stringify(obj, null, 2) + "\n";
  await fs.writeFile(filePath, content, "utf-8");
  console.log(`Arquivo ${filePath} atualizado`);
}

function syncAndReorder(template, target) {
  return Object.entries(template).reduce((acc, [key, value]) => {
    if (key in target) {
      const tVal = target[key];
      if (
        value &&
        typeof value === "object" &&
        !Array.isArray(value) &&
        tVal &&
        typeof tVal === "object" &&
        !Array.isArray(tVal)
      ) {
        acc[key] = syncAndReorder(value, tVal);
      } else {
        acc[key] = tVal;
      }
    } else {
      if (value && typeof value === "object" && !Array.isArray(value)) {
        acc[key] = syncAndReorder(value, {});
      } else {
        acc[key] = "";
      }
    }
    return acc;
  }, {});
}

const main = async () => {
  const loaded = await Promise.all(
    LOCALES.map(async (locale) => {
      const path = join(LOCALES_DIR, `${locale}.json`);
      const data = await readJson(path);
      return [locale, data];
    }),
  );
  if (loaded.some(([, data]) => data === null))
    return void console.log(
      "Não foi possivel carregar um dos arquivos, toda a ação foi cancelada.",
    );

  const data = Object.fromEntries(loaded);
  const base = data[DEFAULT_LOCALE];

  await Promise.all(
    LOCALES.filter((locale) => locale !== DEFAULT_LOCALE).map(async (locale) => {
      const path = join(LOCALES_DIR, `${locale}.json`);
      const result = syncAndReorder(base, data[locale]);
      await writeJson(path, result);
    }),
  );
};

main().catch((err) => {
  console.error("Erro inesperado:", err);
  process.exit(1);
});
