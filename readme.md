# Guia - irn-renamer (Windows)

## 1. Introdução

A ferramenta `renamer` foi criada para simplificar e automatizar o processo de configuração inicial de um novo projeto baseado em um template C#. Ela renomeia pastas, arquivos (incluindo `.sln` e `.csproj`), e substitui ocorrências de texto (como namespaces e referências) dentro dos arquivos, tudo com base nos parâmetros inseridos.

**Público:** Ferramenta exclusiva para desenvolvedores do IRN que utilizam projetos C# base.

**Pré-requisitos:** Windows 10 ou superior.

## 2. Download

Antes de usar, você precisa baixar o executável da ferramenta.

1.  Acesse a página de **Releases** do projeto no GitHub: [irn-renamer](https://github.com/rafaelbatistaroque/irn-renamer/releases)

2.  Localize a versão mais recente ou a desejada.
3.  Na seção **"Assets"** dessa release, encontre o arquivo compilado para Windows 64-bit. O nome deve ser parecido com `irn-renamer.exe`.

![image](https://github.com/user-attachments/assets/ae6f721b-4d92-4ebc-9f3e-dc0c8837ed7f)

4.  Clique no nome do arquivo para fazer o download (geralmente para a sua pasta `Downloads`).

> OBS: Por ser um ficheiro executável, o seu browser pode bloquear o download. Apenas aceite as condições para manter o download. 

## 3. Instalação e Configuração do Ambiente

Para poder chamar o comando `irn-renamer` de qualquer lugar no seu terminal, vamos seguir estes passos:

### 3.1. Criar a pasta tools na raiz do seu ambiente `C:\tools` (pode ser qualquer nome).

Esta pasta servirá como um local central para suas ferramentas de linha de comando. Se ela ainda não existir:

1.  Abra o **Explorador de Arquivos** (tecla `Windows + E`).
2.  Navegue até `Este Computador` > `Disco Local (C:)`.
3.  Clique com o botão direito do mouse numa área vazia, vá em `Novo` > `Pasta`.
4.  Nomeie a pasta como **`tools`**.

### 3.2. Mover e Renomear o Executável

1.  Localize o arquivo que você baixou (ex: `irn-renamer.exe`) na sua pasta `Downloads`.
2.  **Mova** este arquivo para dentro da pasta `C:\tools`.
3.  Dentro de `C:\tools`, **renomeie** o arquivo de `renamer_windows_amd64.exe` para simplesmente **`irn-renamer.exe`**.
![Screenshot 2025-03-28 173558](https://github.com/user-attachments/assets/a8ed2446-375f-443b-9b6e-b86f9ce7e616)

### 3.3. Adicionar `C:\tools` às Variáveis de Ambiente (PATH)

Isso permite que o Windows encontre o `irn-renamer.exe` quando você digitar o comando no terminal.

1.  Pressione a tecla `Windows`, digite `variáveis de ambiente` e clique em **"Editar as variáveis de ambiente do sistema"**.

2.  Na janela "Propriedades do Sistema" que se abre, clique no botão **"Variáveis de Ambiente..."**.
![Screenshot 2025-03-28 171611](https://github.com/user-attachments/assets/9d9d1302-36f6-4d12-85a7-8cdca825c1fa)

3.  Na seção superior ("Variáveis de usuário para [seu usuário]"), localize e selecione a variável `Path`. Clique em **"Editar..."**. (Se a variável `Path` não existir, clique em "Novo..." para criá-la primeiro).
![Screenshot 2025-03-28 171950](https://github.com/user-attachments/assets/25c7bc0c-d609-4057-ac5c-67a1505d13de)

5.  Na janela "Editar a variável de ambiente", clique em **"Novo"** ou **"Procurar"**.
 - Se clicar em **"Novo"**, digite `C:\tools` na nova linha que apareceu e pressione Enter.
 - Se clicar em **"Procurar"**, selecione o diretório que que acabou de criar: Ex: C:/tools.
![Screenshot 2025-03-28 172048](https://github.com/user-attachments/assets/90c60b10-4a99-4c08-8eae-da8676b61124)

6.  Clique em **"OK"** em todas as janelas abertas ("Editar a variável de ambiente", "Variáveis de Ambiente", "Propriedades do Sistema") para salvar as alterações.

### 3.4. Verificação

1.  **Importante:** Feche **todas** as janelas do Prompt de Comando ou PowerShell que estiverem abertas.
2.  Abra uma **NOVA** janela do Prompt de Comando ou PowerShell.
3.  Digite `irn-renamer` e pressione Enter.
4.  Se tudo deu certo, você verá uma mensagem da ferramenta indicando que faltam os argumentos `-old` e `-new`. Isso confirma que o Windows a encontrou! Caso contrário, revise os passos anteriores.

## 4. Utilização do Comando `irn-renamer`

1.  **Navegue até o Projeto:** Abra seu terminal (Prompt/PowerShell) e use `cd` para entrar na pasta **raiz** do projeto C# que você acabou de copiar do template e deseja renomear.
    ```bash
    # Exemplo:
    cd C:\MeusProjetos\ProjetoQueIraModificar
    ```

2.  **Execute o Comando:** Use o comando `irn-renamer` com as flags `-old` e `-new`, especificando o nome que será substituído no parâmetro -old e o novo nome desejado no parâmetro -new.
    ```bash
    irn-renamer -old NomeParaSubstituir" -new NomeNovo
    ```
    * Substitua `"NomeParaSubstituir"` pelo nome usado no seu template (ex: "Company.Template.Service").
    * Substitua `"NomeNovo"` pelo nome desejado para este novo projeto (ex: "Company.Customer.Service").
    * Evite espaços em nomes de projetos/namespaces.

3.  **Confirmação de Segurança:** A ferramenta exibirá um aviso sobre backup e perguntará se você deseja continuar. **Leia o aviso!**
    * Digite `S` ou `Y` (maiúsculo ou minúsculo) e pressione Enter para confirmar e iniciar o processo.
    * Qualquer outra tecla ou apenas Enter abortará a operação.

![Screenshot 2025-03-28 173400](https://github.com/user-attachments/assets/ad30eb25-839c-4beb-b783-0088cb81c145)

4.  **Acompanhe:** A ferramenta listará as alterações que está fazendo (renomeando, atualizando conteúdo).
7.  **Verifique:** Após a conclusão, abra o projeto no Visual Studio ou VS Code, verifique os nomes de arquivos, pastas, `.sln`, `.csproj` e o conteúdo de alguns arquivos `.cs` (namespaces, referências) para garantir que tudo foi alterado conforme esperado. Compile e teste o projeto.

---
*Fim do Guia*