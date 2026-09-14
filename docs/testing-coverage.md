# Testing Coverage

Testing criteria for a passing coverage requirement:

- Line coverage of 80%
- Cognitive complexity of 0
- Have cognitive complexity < 5, but have any coverage

Low cognitive complexity means there are few conditional branches to cover. Tests with cognitive complexity 0 would be covered by invocation.

## Packages

| Status | Package              | Coverage | Cognitive | Lines |
|--------|----------------------|----------|-----------|-------|
| ✅     | .                    | 84.85%   | 24        | 104   |
| ❌     | cmd/lessgo           | 0.00%    | 30        | 209   |
| ❌     | dst                  | 59.56%   | 742       | 2102  |
| ❌     | evaluator            | 65.79%   | 73        | 167   |
| ✅     | examples             | 85.11%   | 15        | 156   |
| ❌     | expression           | 78.33%   | 343       | 1352  |
| ❌     | expression/functions | 73.20%   | 491       | 2220  |
| ❌     | internal/strings     | 63.41%   | 23        | 65    |
| ✅     | renderer             | 82.25%   | 523       | 1250  |

## Functions

| Status | Package              | Function                                   | Coverage | Cognitive |
|--------|----------------------|--------------------------------------------|----------|-----------|
| ❌     | .                    | Handler.ServeHTTP                          | 80.00%   | 14        |
| ✅     |                      | NewHandler                                 | 100.00%  | 0         |
| ✅     |                      | NewMiddleware                              | 100.00%  | 10        |
| ❌     | cmd/lessgo           | astCmd                                     | 0.00%    | 4         |
| ❌     |                      | fmtCmd                                     | 0.00%    | 8         |
| ❌     |                      | generateCmd                                | 0.00%    | 16        |
| ❌     |                      | main                                       | 0.00%    | 2         |
| ✅     |                      | printUsage                                 | 0.00%    | 0         |
| ✅     | dst                  | AssumeNoAllocParser                        | 0.00%    | 0         |
| ✅     |                      | Block.Names                                | 100.00%  | 0         |
| ✅     |                      | Block.Type                                 | 0.00%    | 0         |
| ✅     |                      | BlockVariable.Names                        | 100.00%  | 0         |
| ✅     |                      | BlockVariable.Type                         | 0.00%    | 0         |
| ✅     |                      | Comment.Names                              | 100.00%  | 0         |
| ✅     |                      | Comment.Type                               | 0.00%    | 0         |
| ✅     |                      | Decl.Names                                 | 100.00%  | 0         |
| ✅     |                      | Decl.Type                                  | 0.00%    | 0         |
| ✅     |                      | Each.Names                                 | 100.00%  | 0         |
| ✅     |                      | Each.Type                                  | 0.00%    | 0         |
| ❌     |                      | Formatter.Format                           | 0.00%    | 1         |
| ❌     |                      | Formatter.formatBlock                      | 0.00%    | 7         |
| ❌     |                      | Formatter.formatComment                    | 0.00%    | 2         |
| ❌     |                      | Formatter.formatDecl                       | 0.00%    | 5         |
| ❌     |                      | Formatter.formatEach                       | 0.00%    | 1         |
| ❌     |                      | Formatter.formatMixinCall                  | 0.00%    | 3         |
| ❌     |                      | Formatter.formatNode                       | 0.00%    | 1         |
| ❌     |                      | Formatter.writeIndent                      | 0.00%    | 1         |
| ✅     |                      | Guard.Valid                                | 100.00%  | 1         |
| ✅     |                      | Import.Names                               | 0.00%    | 0         |
| ✅     |                      | Import.Type                                | 0.00%    | 0         |
| ✅     |                      | MixinCall.Names                            | 100.00%  | 0         |
| ✅     |                      | MixinCall.Type                             | 0.00%    | 0         |
| ✅     |                      | NewFormatter                               | 0.00%    | 0         |
| ✅     |                      | NewParser                                  | 100.00%  | 0         |
| ❌     |                      | NewParserForConfig                         | 0.00%    | 1         |
| ❌     |                      | NewParserForConfigWithFS                   | 0.00%    | 1         |
| ✅     |                      | NewParserNoAlloc                           | 100.00%  | 0         |
| ✅     |                      | NewParserNoAllocWithFS                     | 0.00%    | 0         |
| ✅     |                      | NewParserWithFS                            | 100.00%  | 0         |
| ✅     |                      | Parser.Parse                               | 90.91%   | 57        |
| ❌     |                      | Parser.parseBlock                          | 70.87%   | 81        |
| ❌     |                      | Parser.parseBlockVariable                  | 50.91%   | 40        |
| ✅     |                      | Parser.parseDecl                           | 96.15%   | 1         |
| ❌     |                      | Parser.parseEach                           | 75.38%   | 34        |
| ✅     |                      | Parser.parseImport                         | 88.33%   | 9         |
| ✅     |                      | Parser.readMultilineComment                | 52.63%   | 4         |
| ✅     |                      | Parser.scan                                | 100.00%  | 2         |
| ❌     |                      | ParserNoAlloc.Parse                        | 50.91%   | 132       |
| ✅     |                      | ParserNoAlloc.analyzeLinePattern           | 100.00%  | 11        |
| ❌     |                      | ParserNoAlloc.isMixinName                  | 0.00%    | 2         |
| ❌     |                      | ParserNoAlloc.parseArgsNoAlloc             | 0.00%    | 3         |
| ❌     |                      | ParserNoAlloc.parseBlockNoAlloc            | 67.47%   | 49        |
| ❌     |                      | ParserNoAlloc.parseBlockVariableNoAlloc    | 15.91%   | 27        |
| ✅     |                      | ParserNoAlloc.parseDeclNoAlloc             | 92.86%   | 2         |
| ❌     |                      | ParserNoAlloc.parseEachNoAlloc             | 0.00%    | 43        |
| ❌     |                      | ParserNoAlloc.parseImportNoAlloc           | 0.00%    | 4         |
| ❌     |                      | ParserNoAlloc.readMultilineCommentNoAlloc  | 36.36%   | 11        |
| ✅     |                      | ParserNoAlloc.scan                         | 92.31%   | 2         |
| ❌     |                      | Print                                      | 0.00%    | 2         |
| ✅     |                      | SanitizeBytes                              | 99.18%   | 70        |
| ✅     |                      | SanitizeReader                             | 80.00%   | 1         |
| ✅     |                      | UseParser                                  | 0.00%    | 0         |
| ✅     |                      | containsRealBrace                          | 100.00%  | 8         |
| ❌     |                      | findBlockBraces                            | 70.00%   | 18        |
| ✅     |                      | getTrimmed                                 | 100.00%  | 6         |
| ✅     |                      | isDigitFunc                                | 100.00%  | 1         |
| ✅     |                      | isLetterFunc                               | 100.00%  | 3         |
| ❌     |                      | isValidVarName                             | 75.00%   | 7         |
| ✅     |                      | normalizeCommas                            | 100.00%  | 23        |
| ❌     |                      | printNode                                  | 0.00%    | 27        |
| ❌     |                      | splitCommaNoAlloc                          | 0.00%    | 9         |
| ❌     |                      | splitParameterList                         | 67.74%   | 12        |
| ✅     |                      | splitSelectorList                          | 100.00%  | 14        |
| ✅     |                      | trimTrailingWhitespace                     | 100.00%  | 3         |
| ✅     | evaluator            | IsExpression                               | 100.00%  | 3         |
| ❌     |                      | ParseExpression                            | 0.00%    | 8         |
| ❌     |                      | Tokenize                                   | 76.34%   | 57        |
| ❌     |                      | isDigit                                    | 0.00%    | 1         |
| ❌     |                      | isValueChar                                | 0.00%    | 4         |
| ✅     | examples             | Example1_Middleware                        | 100.00%  | 0         |
| ✅     |                      | Example1_MiddlewareWithChain               | 100.00%  | 0         |
| ✅     |                      | Example2_CustomHandler                     | 100.00%  | 0         |
| ✅     |                      | Example2_MuxWithHandler                    | 100.00%  | 0         |
| ❌     |                      | LessCompilerHandler.ServeHTTP              | 65.22%   | 6         |
| ✅     |                      | LessCompilerHandler.compileLess            | 88.89%   | 3         |
| ✅     |                      | LessCompilerHandler.isValidLessFile        | 66.67%   | 3         |
| ✅     |                      | NewLessCompilerHandler                     | 85.71%   | 3         |
| ✅     | expression           | Call                                       | 75.00%   | 1         |
| ❌     |                      | Color.Darken                               | 0.00%    | 2         |
| ❌     |                      | Color.Desaturate                           | 0.00%    | 2         |
| ❌     |                      | Color.Lighten                              | 0.00%    | 2         |
| ❌     |                      | Color.Saturate                             | 0.00%    | 2         |
| ❌     |                      | Color.Spin                                 | 0.00%    | 4         |
| ✅     |                      | Color.String                               | 88.89%   | 6         |
| ✅     |                      | Evaluator.Eval                             | 100.00%  | 0         |
| ✅     |                      | Evaluator.SetVariable                      | 100.00%  | 0         |
| ❌     |                      | Evaluator.evalExpression                   | 78.95%   | 18        |
| ✅     |                      | Evaluator.evalFunctionCall                 | 92.00%   | 7         |
| ❌     |                      | Evaluator.evaluateEmbeddedFunctions        | 0.00%    | 3         |
| ❌     |                      | Evaluator.extractFunctions                 | 0.00%    | 34        |
| ✅     |                      | Evaluator.parseAddSub                      | 86.96%   | 9         |
| ✅     |                      | Evaluator.parseMulDiv                      | 83.33%   | 12        |
| ✅     |                      | Evaluator.parseValue                       | 60.00%   | 5         |
| ✅     |                      | Evaluator.substituteVariables              | 71.43%   | 5         |
| ✅     |                      | FuncMap                                    | 0.00%    | 0         |
| ✅     |                      | GetRegisteredFunctionNames                 | 100.00%  | 0         |
| ✅     |                      | IsFunctionCall                             | 100.00%  | 2         |
| ✅     |                      | IsRegisteredFunction                       | 100.00%  | 0         |
| ❌     |                      | List.Extract                               | 0.00%    | 2         |
| ✅     |                      | List.Length                                | 0.00%    | 0         |
| ✅     |                      | List.String                                | 0.00%    | 0         |
| ✅     |                      | NewColorValue                              | 100.00%  | 0         |
| ✅     |                      | NewEvaluator                               | 88.89%   | 3         |
| ✅     |                      | NewList                                    | 0.00%    | 0         |
| ✅     |                      | NewValue                                   | 100.00%  | 0         |
| ✅     |                      | Parse                                      | 96.88%   | 13        |
| ✅     |                      | ParseColor                                 | 90.00%   | 5         |
| ✅     |                      | ParseFunctionCall                          | 89.29%   | 11        |
| ❌     |                      | ParseList                                  | 0.00%    | 15        |
| ✅     |                      | Value.Add                                  | 100.00%  | 5         |
| ✅     |                      | Value.Divide                               | 88.89%   | 4         |
| ✅     |                      | Value.Multiply                             | 100.00%  | 3         |
| ✅     |                      | Value.String                               | 96.00%   | 9         |
| ✅     |                      | Value.Subtract                             | 71.43%   | 5         |
| ✅     |                      | absFloat                                   | 100.00%  | 1         |
| ✅     |                      | containsOperator                           | 100.00%  | 21        |
| ✅     |                      | fmod                                       | 100.00%  | 0         |
| ✅     |                      | hslToRGB                                   | 91.30%   | 4         |
| ✅     |                      | init                                       | 100.00%  | 1         |
| ✅     |                      | isColorLike                                | 100.00%  | 1         |
| ✅     |                      | isDigit                                    | 100.00%  | 1         |
| ✅     |                      | isIdentifierChar                           | 100.00%  | 4         |
| ✅     |                      | isOperator                                 | 100.00%  | 3         |
| ❌     |                      | max                                        | 0.00%    | 1         |
| ❌     |                      | maxFloat                                   | 0.00%    | 4         |
| ❌     |                      | min                                        | 0.00%    | 1         |
| ❌     |                      | minFloat                                   | 0.00%    | 4         |
| ✅     |                      | parseHSLColor                              | 85.37%   | 17        |
| ✅     |                      | parseHexColor                              | 88.24%   | 15        |
| ✅     |                      | parseRGBColor                              | 80.56%   | 17        |
| ✅     |                      | register                                   | 86.67%   | 30        |
| ✅     |                      | registerFunctions                          | 100.00%  | 0         |
| ❌     |                      | rgbToHSL                                   | 0.00%    | 5         |
| ✅     |                      | splitArgs                                  | 100.00%  | 16        |
| ✅     |                      | splitByOperator                            | 85.19%   | 8         |
| ✅     |                      | trimFloat                                  | 100.00%  | 0         |
| ✅     | expression/functions | ARGB                                       | 87.50%   | 1         |
| ✅     |                      | Abs                                        | 100.00%  | 0         |
| ✅     |                      | Acos                                       | 0.00%    | 0         |
| ✅     |                      | Alpha                                      | 75.00%   | 1         |
| ✅     |                      | Asin                                       | 0.00%    | 0         |
| ✅     |                      | Atan                                       | 0.00%    | 0         |
| ✅     |                      | Average                                    | 87.50%   | 2         |
| ✅     |                      | Blue                                       | 75.00%   | 1         |
| ✅     |                      | Boolean                                    | 100.00%  | 1         |
| ✅     |                      | Ceil                                       | 100.00%  | 0         |
| ✅     |                      | ClearImageDimCache                         | 0.00%    | 0         |
| ✅     |                      | Color.Darken                               | 100.00%  | 0         |
| ✅     |                      | Color.Desaturate                           | 100.00%  | 0         |
| ✅     |                      | Color.Greyscale                            | 100.00%  | 0         |
| ✅     |                      | Color.Lighten                              | 100.00%  | 0         |
| ✅     |                      | Color.Luma                                 | 100.00%  | 0         |
| ✅     |                      | Color.Mix                                  | 100.00%  | 0         |
| ✅     |                      | Color.Saturate                             | 100.00%  | 0         |
| ✅     |                      | Color.Spin                                 | 80.00%   | 1         |
| ✅     |                      | Color.ToHSL                                | 100.00%  | 5         |
| ✅     |                      | Color.ToHSV                                | 90.48%   | 5         |
| ✅     |                      | Color.ToHex                                | 81.82%   | 1         |
| ❌     |                      | Color.ToRGB                                | 0.00%    | 1         |
| ✅     |                      | ColorFunction                              | 85.71%   | 6         |
| ✅     |                      | Contrast                                   | 93.75%   | 10        |
| ✅     |                      | Convert                                    | 84.09%   | 10        |
| ✅     |                      | Cos                                        | 100.00%  | 0         |
| ✅     |                      | Darken                                     | 83.33%   | 1         |
| ✅     |                      | Desaturate                                 | 83.33%   | 1         |
| ✅     |                      | Difference                                 | 87.50%   | 2         |
| ✅     |                      | E                                          | 100.00%  | 5         |
| ✅     |                      | Escape                                     | 100.00%  | 5         |
| ❌     |                      | EvaluateExpression                         | 0.00%    | 10        |
| ✅     |                      | Exclusion                                  | 91.67%   | 2         |
| ❌     |                      | Extract                                    | 44.00%   | 11        |
| ✅     |                      | Fade                                       | 90.91%   | 1         |
| ✅     |                      | Fadein                                     | 88.89%   | 1         |
| ✅     |                      | Fadeout                                    | 88.89%   | 1         |
| ✅     |                      | Floor                                      | 100.00%  | 0         |
| ✅     |                      | Format                                     | 95.83%   | 22        |
| ✅     |                      | GetUnit                                    | 100.00%  | 0         |
| ✅     |                      | Green                                      | 75.00%   | 1         |
| ✅     |                      | Greyscale                                  | 80.00%   | 1         |
| ✅     |                      | HSL                                        | 94.12%   | 1         |
| ✅     |                      | HSLA                                       | 95.24%   | 1         |
| ❌     |                      | HSLToColor                                 | 76.00%   | 10        |
| ✅     |                      | HSV                                        | 100.00%  | 0         |
| ✅     |                      | HSVA                                       | 100.00%  | 0         |
| ✅     |                      | HSVHue                                     | 88.89%   | 1         |
| ✅     |                      | HSVSaturation                              | 80.00%   | 1         |
| ❌     |                      | HSVToColor                                 | 68.00%   | 10        |
| ✅     |                      | HSVValue                                   | 80.00%   | 1         |
| ✅     |                      | Hardlight                                  | 85.71%   | 4         |
| ✅     |                      | Hue                                        | 80.00%   | 1         |
| ❌     |                      | If                                         | 16.67%   | 45        |
| ✅     |                      | ImageHeight                                | 77.78%   | 3         |
| ✅     |                      | ImageSize                                  | 77.78%   | 3         |
| ✅     |                      | ImageWidth                                 | 77.78%   | 3         |
| ✅     |                      | IsColor                                    | 83.33%   | 19        |
| ✅     |                      | IsColorFunction                            | 100.00%  | 1         |
| ✅     |                      | IsDefined                                  | 0.00%    | 0         |
| ✅     |                      | IsEm                                       | 100.00%  | 1         |
| ✅     |                      | IsEmFunction                               | 100.00%  | 1         |
| ❌     |                      | IsKeyword                                  | 58.33%   | 8         |
| ✅     |                      | IsKeywordFunction                          | 66.67%   | 1         |
| ❌     |                      | IsList                                     | 0.00%    | 1         |
| ❌     |                      | IsListFunction                             | 0.00%    | 1         |
| ✅     |                      | IsNumber                                   | 83.33%   | 11        |
| ✅     |                      | IsNumberFunction                           | 100.00%  | 1         |
| ✅     |                      | IsPercentage                               | 100.00%  | 1         |
| ✅     |                      | IsPercentageFunction                       | 100.00%  | 1         |
| ✅     |                      | IsPixel                                    | 100.00%  | 1         |
| ✅     |                      | IsPixelFunction                            | 100.00%  | 1         |
| ✅     |                      | IsRuleset                                  | 83.33%   | 3         |
| ❌     |                      | IsRulesetFunction                          | 0.00%    | 1         |
| ✅     |                      | IsString                                   | 75.00%   | 4         |
| ✅     |                      | IsStringFunction                           | 100.00%  | 1         |
| ✅     |                      | IsURL                                      | 100.00%  | 1         |
| ✅     |                      | IsURLFunction                              | 100.00%  | 1         |
| ✅     |                      | IsUnit                                     | 100.00%  | 5         |
| ✅     |                      | IsUnitFunction                             | 100.00%  | 1         |
| ✅     |                      | Length                                     | 75.00%   | 5         |
| ✅     |                      | Lighten                                    | 83.33%   | 1         |
| ✅     |                      | Lightness                                  | 80.00%   | 1         |
| ✅     |                      | LumaFunction                               | 85.71%   | 1         |
| ✅     |                      | Luminance                                  | 95.24%   | 1         |
| ✅     |                      | Max                                        | 92.31%   | 4         |
| ✅     |                      | Min                                        | 92.31%   | 4         |
| ❌     |                      | Mix                                        | 76.92%   | 8         |
| ✅     |                      | Mod                                        | 88.89%   | 1         |
| ✅     |                      | Multiply                                   | 87.50%   | 2         |
| ✅     |                      | Negation                                   | 87.50%   | 2         |
| ✅     |                      | Overlay                                    | 92.86%   | 4         |
| ✅     |                      | ParseColor                                 | 91.67%   | 4         |
| ❌     |                      | ParseHSL                                   | 80.00%   | 16        |
| ✅     |                      | ParseHex                                   | 67.86%   | 1         |
| ❌     |                      | ParseRGB                                   | 67.86%   | 16        |
| ✅     |                      | Percentage                                 | 100.00%  | 0         |
| ✅     |                      | Pi                                         | 100.00%  | 1         |
| ✅     |                      | Pow                                        | 100.00%  | 0         |
| ✅     |                      | RGB                                        | 100.00%  | 0         |
| ✅     |                      | Range                                      | 83.33%   | 14        |
| ✅     |                      | Red                                        | 75.00%   | 1         |
| ✅     |                      | Replace                                    | 93.18%   | 37        |
| ✅     |                      | Round                                      | 100.00%  | 0         |
| ✅     |                      | Saturate                                   | 83.33%   | 1         |
| ✅     |                      | Saturation                                 | 80.00%   | 1         |
| ✅     |                      | Screen                                     | 94.74%   | 2         |
| ✅     |                      | Shade                                      | 90.91%   | 1         |
| ✅     |                      | Sin                                        | 100.00%  | 0         |
| ❌     |                      | Softlight                                  | 66.67%   | 7         |
| ✅     |                      | Spin                                       | 83.33%   | 1         |
| ✅     |                      | Sqrt                                       | 100.00%  | 0         |
| ✅     |                      | Tan                                        | 100.00%  | 0         |
| ✅     |                      | Tint                                       | 90.91%   | 1         |
| ❌     |                      | Unit                                       | 53.85%   | 16        |
| ❌     |                      | evaluateBinaryOp                           | 0.00%    | 9         |
| ❌     |                      | evaluateExpressionSimple                   | 0.00%    | 3         |
| ❌     |                      | evaluateSimpleComparison                   | 0.00%    | 19        |
| ✅     |                      | extractUnit                                | 100.00%  | 4         |
| ✅     |                      | formatColor                                | 83.33%   | 1         |
| ✅     |                      | formatNumber                               | 100.00%  | 0         |
| ✅     |                      | formatNumberWithUnit                       | 100.00%  | 5         |
| ✅     |                      | gammaCorrect                               | 100.00%  | 1         |
| ✅     |                      | getImageDimensions                         | 93.10%   | 5         |
| ✅     |                      | hasRegexMetacharacters                     | 80.00%   | 3         |
| ❌     |                      | isPartOfNumber                             | 0.00%    | 3         |
| ✅     |                      | parseAlpha                                 | 100.00%  | 0         |
| ✅     |                      | parseChannelNumber                         | 100.00%  | 0         |
| ✅     |                      | parseHexByte                               | 100.00%  | 0         |
| ✅     |                      | parseHexDigit                              | 100.00%  | 0         |
| ✅     |                      | parseNumber                                | 100.00%  | 4         |
| ✅     |                      | parseNumberWithUnits                       | 92.31%   | 7         |
| ❌     |                      | preprocessComparisonExpr                   | 0.00%    | 1         |
| ❌     |                      | regexReplaceFirst                          | 0.00%    | 1         |
| ✅     |                      | roundHSLValue                              | 100.00%  | 0         |
| ✅     |                      | stringReplace                              | 66.67%   | 2         |
| ❌     |                      | tryEvaluateSimpleExpr                      | 38.71%   | 10        |
| ❌     | internal/strings     | SplitByteNoAlloc                           | 0.00%    | 9         |
| ✅     |                      | SplitCommaNoAlloc                          | 100.00%  | 9         |
| ✅     |                      | TrimSpace                                  | 100.00%  | 4         |
| ✅     |                      | isSpace                                    | 100.00%  | 1         |
| ✅     | renderer             | NewRenderer                                | 100.00%  | 0         |
| ✅     |                      | NewResolver                                | 100.00%  | 0         |
| ✅     |                      | NewStack                                   | 100.00%  | 0         |
| ✅     |                      | NodeContext.Depth                          | 100.00%  | 0         |
| ✅     |                      | Renderer.Render                            | 100.00%  | 0         |
| ✅     |                      | Renderer.RenderWithBaseDir                 | 97.06%   | 2         |
| ✅     |                      | Renderer.collectBlockVariables             | 100.00%  | 3         |
| ✅     |                      | Renderer.collectMixinsAndExtends           | 100.00%  | 0         |
| ✅     |                      | Renderer.collectMixinsAndExtendsWithPrefix | 100.00%  | 65        |
| ✅     |                      | Renderer.evaluateGuard                     | 83.33%   | 38        |
| ✅     |                      | Renderer.parseExtendSelectors              | 38.46%   | 1         |
| ✅     |                      | Renderer.renderBlock                       | 91.80%   | 103       |
| ✅     |                      | Renderer.renderComment                     | 100.00%  | 1         |
| ✅     |                      | Renderer.renderDecl                        | 95.92%   | 11        |
| ✅     |                      | Renderer.renderEach                        | 88.89%   | 8         |
| ✅     |                      | Renderer.renderImport                      | 0.00%    | 0         |
| ✅     |                      | Renderer.renderMediaQueriesForSelector     | 96.33%   | 8         |
| ❌     |                      | Renderer.renderMixinCall                   | 75.00%   | 34        |
| ✅     |                      | Renderer.renderNode                        | 83.33%   | 2         |
| ❌     |                      | Renderer.renderNodes                       | 68.75%   | 18        |
| ❌     |                      | Renderer.renderTopLevelMediaBlock          | 0.00%    | 3         |
| ✅     |                      | Renderer.writeIndent                       | 100.00%  | 1         |
| ✅     |                      | Resolver.InterpolateVariables              | 80.00%   | 2         |
| ✅     |                      | Resolver.ResolveValue                      | 89.74%   | 38        |
| ❌     |                      | Resolver.containsOperator                  | 0.00%    | 49        |
| ✅     |                      | Resolver.evaluateEmbeddedFunctions         | 100.00%  | 3         |
| ✅     |                      | Resolver.extractFunctionsFromValue         | 98.18%   | 81        |
| ✅     |                      | Resolver.substituteVariables               | 86.67%   | 11        |
| ✅     |                      | Stack.All                                  | 100.00%  | 3         |
| ❌     |                      | Stack.Any                                  | 0.00%    | 1         |
| ✅     |                      | Stack.Depth                                | 100.00%  | 0         |
| ✅     |                      | Stack.Get                                  | 100.00%  | 3         |
| ❌     |                      | Stack.GetGlobal                            | 0.00%    | 1         |
| ✅     |                      | Stack.Pop                                  | 100.00%  | 1         |
| ✅     |                      | Stack.Push                                 | 100.00%  | 0         |
| ✅     |                      | Stack.Set                                  | 66.67%   | 1         |
| ✅     |                      | Stack.SetGlobal                            | 100.00%  | 1         |
| ✅     |                      | contains                                   | 75.00%   | 3         |
| ✅     |                      | initEvaluableFuncNames                     | 100.00%  | 4         |
| ✅     |                      | isCSSOnlyFunction                          | 100.00%  | 3         |
| ❌     |                      | isLikeList                                 | 0.00%    | 2         |
| ❌     |                      | isValueChar                                | 0.00%    | 4         |
| ✅     |                      | isVarChar                                  | 100.00%  | 4         |
| ✅     |                      | normalizeSelectorSpacing                   | 95.65%   | 1         |
| ✅     |                      | parseNumberForGuard                        | 100.00%  | 7         |
| ✅     |                      | selector                                   | 100.00%  | 2         |
