#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { UguisuStack } from '../lib/uguisu-stack';

const app = new cdk.App();
new UguisuStack(app, 'UguisuStack');
