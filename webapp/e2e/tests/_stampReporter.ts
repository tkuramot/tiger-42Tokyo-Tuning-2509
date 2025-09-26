import {
  Reporter,
  FullConfig,
  Suite,
  TestCase,
  TestResult,
  FullResult,
} from "@playwright/test/reporter";
import * as fs from "fs";
import * as path from "path";
import { execSync } from "child_process";

class StampReporter implements Reporter {
  private allTestsPassed = true;
  private hasErrors = false;
  private testCount = 0;
  private tokensDir = path.join(process.cwd(), "tokens");

  onBegin(config: FullConfig, suite: Suite) {
    // tokens ディレクトリを作成
    if (!fs.existsSync(this.tokensDir)) {
      fs.mkdirSync(this.tokensDir, { recursive: true });
    }

    // テスト総数をカウント
    this.testCount = this.countTests(suite);
    console.log(`📝 実行予定テスト数: ${this.testCount}`);
  }

  private countTests(suite: Suite): number {
    let count = 0;
    for (const test of suite.allTests()) {
      count++;
    }
    return count;
  }

  onTestEnd(test: TestCase, result: TestResult) {
    console.log(`🧪 テスト: ${test.title} -> ${result.status}`);

    // テスト失敗時にフラグを更新
    if (result.status !== "passed") {
      this.allTestsPassed = false;
    }
  }

  onError(error: any) {
    console.error("❌ Playwright実行エラー:", error);
    this.hasErrors = true;
    this.allTestsPassed = false;
  }

  onEnd(result: FullResult) {
    console.log(`📊 テスト結果: status=${result.status}`);

    // エラーがある、または実際にテストが実行されていない場合は失敗扱い
    if (this.hasErrors || result.status !== "passed" || this.testCount === 0) {
      console.log(
        "❌ E2Eテストでエラーが発生したか、テストが実行されませんでした。証跡は生成されません。"
      );
      process.exit(1); // 明示的に失敗で終了
      return;
    }

    if (this.allTestsPassed) {
      console.log("✅ 全てのE2Eテストが成功しました。証跡を生成中...");
      this.generateProof();
    } else {
      console.log("❌ E2Eテストに失敗があります。証跡は生成されません。");
      process.exit(1); // 明示的に失敗で終了
    }
  }

  private generateProof() {
    try {
      const timestamp = Math.floor(Date.now() / 1000);

      // 証跡データを生成
      const proofData = {
        timestamp: timestamp,
        test_passed: true,
        test_run_id: `e2e_${timestamp}_${Math.random()
          .toString(36)
          .substring(2, 11)}`,
        test_count: this.testCount,
      };

      const proofJson = JSON.stringify(proofData);

      // JSONファイルを保存
      const jsonPath = path.join(this.tokensDir, "e2e_proof.json");
      fs.writeFileSync(jsonPath, proofJson);

      // 署名を生成（秘密鍵がコンテナ内にある場合）
      if (fs.existsSync("/etc/e2e-private.key")) {
        const sigPath = path.join(this.tokensDir, "e2e_proof.sig");
        // シェルインジェクション対策でファイル経由
        const tempFile = path.join(this.tokensDir, "temp_proof.txt");
        fs.writeFileSync(tempFile, proofJson);
        execSync(
          `openssl dgst -sha256 -sign /etc/e2e-private.key -binary "${tempFile}" | base64 > "${sigPath}"`
        );
        fs.unlinkSync(tempFile); // 一時ファイル削除
        console.log("🔐 署名付き証跡を生成しました");
      } else {
        // 署名なしの場合は成功マーカーのみ
        const successPath = path.join(this.tokensDir, "test_success.marker");
        fs.writeFileSync(successPath, proofJson);
        console.log("📝 テスト成功証跡を生成しました");
      }

      // 既存の証跡ファイル形式も生成（後方互換性のため）
      const binPath = path.join(this.tokensDir, `${timestamp}.bin`);
      fs.writeFileSync(binPath, Buffer.from(proofJson));
    } catch (error) {
      console.error("証跡生成に失敗:", error);
      process.exit(1);
    }
  }
}

export default StampReporter;
