import { ZxcvbnFactory } from '@zxcvbn-ts/core';
import { adjacencyGraphs, dictionary as commonDictionary } from '@zxcvbn-ts/language-common';
import { dictionary, translations } from '@zxcvbn-ts/language-en';

const estimator = new ZxcvbnFactory({
  graphs: adjacencyGraphs,
  dictionary: { ...commonDictionary, ...dictionary },
  translations,
});

export function passwordStrength(password: string) {
  const score = estimator.check(password.slice(0, 256)).score;
  const level = password.length < 8 || score < 2 ? 1 : score < 4 ? 2 : 3;
  return { label: ['Weak', 'Medium', 'Strong'][level - 1], level };
}
